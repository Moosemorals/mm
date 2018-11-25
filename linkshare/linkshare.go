package linkshare

type linkshare struct{}

func (ls *linkshare) OnMessage(h *Hub, msg Message, r Reply) {

	r(msg)
}
