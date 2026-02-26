package chat

// Internal UI messages (sharkfin messages are in pkg/sharkfin)

type ChannelSelectedMsg struct {
	Name string
}

type ModalCloseMsg struct{}
