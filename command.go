package main

// CmdMsg is a json response object from IoT containing an asset command.
type CmdMsg struct {
	Cmd string `json:"cmd"`
}

func (a *Asset) readCmds() error {
	go func() {
		for {
			msg := <-a.cmdMsg
			customOut("cmd received: "+msg.Cmd, nil)
		}
	}()
	customOut("command channel open", nil)
	return nil
}
