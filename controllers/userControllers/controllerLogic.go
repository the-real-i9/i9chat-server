package userControllers

import (
	"fmt"
	user "i9chat/models/userModel"
	"i9chat/services/messageBrokerService"

	"github.com/jackc/pgx/v5/pgtype"
)

func switchMyPresence(clientUserId int, presence string, lastSeen pgtype.Timestamp) error {
	userDMChatPartnersIdList, err := user.SwitchPresence(clientUserId, presence, lastSeen)
	if err != nil {
		return err
	}

	go func(recepientIds []*int) {
		// "recepients" are: all users to whom you are a DMChat partner
		for _, rId := range recepientIds {
			rId := *rId
			go messageBrokerService.PostMessage(fmt.Sprintf("user-%d", rId), messageBrokerService.Message{
				Event: "user presence changed",
				Data: map[string]any{
					"userId":   clientUserId,
					"presence": presence,
					"lastSeen": lastSeen,
				},
			}, false)
		}
	}(userDMChatPartnersIdList)

	return nil
}
