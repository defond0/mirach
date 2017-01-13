package main

func (a *Asset) readCmds() error {
	go func() {
		msg := <-a.cmdMsg
		customOut(msg, nil)
	}()
	return nil
}
