package client

// WaitForExit blocks until the connection is closed
func (c *Client) WaitForExit() {
	<-c.DoneChan()
}

// DoneChan returns a channel that will be closed when the connection is closed
func (c *Client) DoneChan() <-chan struct{} {
	return c.connection.Done()
}

// Stop stops the bot, quitting with the given message if possible
func (c *Client) Stop(message string) {
	c.connection.Stop(message)
}

/*
Util functions
Join
Part
Quit
Kick
Nick
Privmsg
Notice

these should be ISUPPORT aware where possible, and so on

*/

// func (c *Client) Join(channels ...string) {}

// func (c *Client) Part(channels ...string) {}

// func (c *Client) Nick(newNickname string) {}
