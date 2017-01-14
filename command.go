package main

// CmdMsg is a json response object from IoT containing an asset command.
type CmdMsg struct {
	Cmd string `json:"cmd"`
}

func (a *Asset) readCmds() error {
	go func() {
		msg := <-a.cmdMsg
		customOut(msg, nil)
	}()
	return nil
}
