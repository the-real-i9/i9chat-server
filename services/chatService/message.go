package chatservice

type Message interface {
	React()
	UpdateDeliveryStatus()
}

type DMMessage struct {
	DMChatId int
	Id       int
}

func (dmm DMMessage) React() {

}

func (dmm DMMessage) UpdateDeliveryStatus() {

}

type GroupMessage struct {
	GroupChatId int
	Id          int
}

func (gpm GroupMessage) React() {

}

func (gpm GroupMessage) AddToDeliveries() {

}

func (gpm GroupMessage) AddToReaders() {

}
