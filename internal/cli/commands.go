package cli

import "errors"

var errNotImplemented = errors.New("not implemented yet")

func cmdNext(args []string) error     { return errNotImplemented }
func cmdStart(args []string) error    { return errNotImplemented }
func cmdPause(args []string) error    { return errNotImplemented }
func cmdComplete(args []string) error { return errNotImplemented }
func cmdCurrent(args []string) error  { return errNotImplemented }
func cmdAdd(args []string) error      { return errNotImplemented }
func cmdList(args []string) error     { return errNotImplemented }
func cmdWaybar(args []string) error   { return errNotImplemented }
