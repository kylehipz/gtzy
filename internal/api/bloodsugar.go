package api

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"gtzy/internal/meter"
	"gtzy/internal/models"
	"gtzy/internal/store"
)

func (s *Server) registerBloodSugarRoutes(r chi.Router) {
	bs := &store.BloodSugarStore{DB: s.DB}

	r.Get("/bloodsugar", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		from, to := q.Get("from"), q.Get("to")
		// Default to the last 30 days when no range is given.
		if from == "" && to == "" {
			from = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		}
		readings, err := bs.List(from, to)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, readings)
	})

	r.Post("/bloodsugar", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ValueMgdl float64 `json:"value_mgdl"`
			TakenAt   string  `json:"taken_at"`
			MealTag   string  `json:"meal_tag"`
			Notes     string  `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid body")
			return
		}
		if body.ValueMgdl <= 0 {
			writeErr(w, http.StatusBadRequest, "value_mgdl must be greater than 0")
			return
		}
		if body.TakenAt == "" {
			body.TakenAt = store.NowUTC()
		}
		reading, err := bs.Create(store.BloodSugarInput{
			ValueMgdl: body.ValueMgdl,
			TakenAt:   body.TakenAt,
			MealTag:   body.MealTag,
			Notes:     body.Notes,
			Source:    "manual",
		})
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, reading)
	})

	r.Patch("/bloodsugar/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			ValueMgdl *float64 `json:"value_mgdl"`
			TakenAt   *string  `json:"taken_at"`
			MealTag   *string  `json:"meal_tag"`
			Notes     *string  `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid body")
			return
		}
		reading, err := bs.Update(id, store.BloodSugarPatch{
			ValueMgdl: body.ValueMgdl,
			TakenAt:   body.TakenAt,
			MealTag:   body.MealTag,
			Notes:     body.Notes,
		})
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, reading)
	})

	r.Delete("/bloodsugar/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := bs.Delete(id); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	// Sync pulls new records off the paired Bluetooth glucose meter. Runs
	// server-side because gtzy serve runs on the machine with the BLE adapter.
	r.Post("/bloodsugar/sync", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()

		fetched, synced, err := meter.SyncInto(ctx, s.DB)
		if err != nil {
			writeErr(w, http.StatusBadGateway, "meter sync failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"synced": synced, "fetched": fetched})
	})
}

// bloodSugarStats are the aggregate figures fed to the AI summary prompt.
type bloodSugarStats struct {
	Count        int
	Mean         float64
	Min          float64
	Max          float64
	StdDev       float64
	EstA1C       float64 // estimated A1C %, from the ADAG eAG formula
	InRangePct   float64 // % of readings 70-180 mg/dL
	LowPct       float64 // % < 70 mg/dL
	HighPct      float64 // % > 180 mg/dL
	MealTagMeans map[string]float64
}

// computeBloodSugarStats derives summary statistics from a set of readings.
func computeBloodSugarStats(readings []models.BloodSugarReading) bloodSugarStats {
	st := bloodSugarStats{MealTagMeans: map[string]float64{}}
	st.Count = len(readings)
	if st.Count == 0 {
		return st
	}

	var sum, low, high, inRange float64
	st.Min = readings[0].ValueMgdl
	st.Max = readings[0].ValueMgdl
	tagSum := map[string]float64{}
	tagCount := map[string]int{}
	for _, r := range readings {
		v := r.ValueMgdl
		sum += v
		st.Min = math.Min(st.Min, v)
		st.Max = math.Max(st.Max, v)
		switch {
		case v < 70:
			low++
		case v > 180:
			high++
		default:
			inRange++
		}
		tag := r.MealTag
		if tag == "" {
			tag = "untagged"
		}
		tagSum[tag] += v
		tagCount[tag]++
	}
	n := float64(st.Count)
	st.Mean = sum / n

	var variance float64
	for _, r := range readings {
		d := r.ValueMgdl - st.Mean
		variance += d * d
	}
	st.StdDev = math.Sqrt(variance / n)

	// ADAG estimated A1C from mean glucose (mg/dL).
	st.EstA1C = (st.Mean + 46.7) / 28.7
	st.InRangePct = inRange / n * 100
	st.LowPct = low / n * 100
	st.HighPct = high / n * 100
	for tag, s := range tagSum {
		st.MealTagMeans[tag] = s / float64(tagCount[tag])
	}
	return st
}
