package chatmodel

type Message interface {
	CreateReaction()
}

type DMMessage struct {
	DMChatId int
	Id       int
}

func (dmm DMMessage) UpdateDeliveryStatus() {

}

func (dmm DMMessage) CreateReaction() {

}

type GroupMessage struct {
	GroupChatId int
	Id          int
}

func (gpm GroupMessage) UpdateDeliveries() {

}

func (gpm GroupMessage) UpdateReaders() {

}

func (gpm GroupMessage) CreateReaction() {

}
