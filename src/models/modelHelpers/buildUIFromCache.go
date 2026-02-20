package modelHelpers

import (
	"context"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/cache"
	"i9chat/src/services/cloudStorageService"
)

func BuildUserSnippetUIFromCache(ctx context.Context, username string) (userSnippetUI UITypes.UserSnippet, err error) {
	nilVal := UITypes.UserSnippet{}

	userSnippetUI, err = cache.GetUser[UITypes.UserSnippet](ctx, username)
	if err != nil {
		return nilVal, err
	}

	userSnippetUI.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(userSnippetUI.ProfilePicUrl)

	return userSnippetUI, nil
}

func BuildUserProfileUIFromCache(ctx context.Context, username string) (userProfileUI UITypes.UserProfile, err error) {
	nilVal := UITypes.UserProfile{}

	userProfileUI, err = cache.GetUser[UITypes.UserProfile](ctx, username)
	if err != nil {
		return nilVal, err
	}

	userProfileUI.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(userProfileUI.ProfilePicUrl)

	return userProfileUI, nil
}

func BuildGroupInfoUIFromCache(ctx context.Context, groupId string) (groupInfoUI UITypes.GroupInfo, err error) {
	nilVal := UITypes.GroupInfo{}

	groupInfoUI, err = cache.GetGroup[UITypes.GroupInfo](ctx, groupId)
	if err != nil {
		return nilVal, err
	}

	groupInfoUI.PictureUrl = cloudStorageService.GroupPicCloudNameToUrl(groupInfoUI.PictureUrl)

	groupInfoUI.MembersCount, err = cache.GetGroupMembersCount(ctx, groupId)
	if err != nil {
		return nilVal, err
	}

	groupInfoUI.OnlineMembersCount, err = cache.GetGroupOnlineMembersCount(ctx, groupId)
	if err != nil {
		return nilVal, err
	}

	return groupInfoUI, nil
}

func buildGroupMemberSnippetUIFromCache(ctx context.Context, muser string) (gmemSnippetUI UITypes.GroupMemberSnippet, err error) {
	nilVal := UITypes.GroupMemberSnippet{}

	gmemSnippetUI, err = cache.GetUser[UITypes.GroupMemberSnippet](ctx, muser)
	if err != nil {
		return nilVal, err
	}

	gmemSnippetUI.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(gmemSnippetUI.ProfilePicUrl)

	return gmemSnippetUI, nil
}

func buildChatSnippetUIFromCache(ctx context.Context, clientUsername, chatIdent string) (chatSnippetUI UITypes.ChatSnippet, err error) {
	nilVal := UITypes.ChatSnippet{}

	chatSnippetUI, err = cache.GetChat[UITypes.ChatSnippet](ctx, clientUsername, chatIdent)
	if err != nil {
		return nilVal, err
	}

	switch chatSnippetUI.Type {
	case "direct":
		csuipu, err := cache.GetUser[UITypes.ChatPartnerUser](ctx, chatSnippetUI.PartnerUser.(string))
		if err != nil {
			return nilVal, err
		}

		csuipu.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(csuipu.ProfilePicUrl)

		chatSnippetUI.PartnerUser = csuipu
	case "group":
		csuig, err := cache.GetGroup[UITypes.ChatGroup](ctx, chatSnippetUI.Group.(string))
		if err != nil {
			return nilVal, err
		}

		csuig.PictureUrl = cloudStorageService.GroupPicCloudNameToUrl(csuig.PictureUrl)

		chatSnippetUI.Group = csuig
	}

	chatSnippetUI.UnreadMC, err = cache.GetChatUnreadMsgsCount(ctx, clientUsername, chatIdent)
	if err != nil {
		return nilVal, err
	}

	return chatSnippetUI, nil
}

func buildCHEUIFromCache(ctx context.Context, CHEId, chatType string) (CHEUI UITypes.ChatHistoryEntry, err error) {
	nilVal := UITypes.ChatHistoryEntry{}

	switch chatType {
	case "direct":
		CHEUI, err = cache.GetDirectChatHistoryEntry[UITypes.ChatHistoryEntry](ctx, CHEId)
		if err != nil {
			return nilVal, err
		}
	case "group":
		CHEUI, err = cache.GetGroupChatHistoryEntry[UITypes.ChatHistoryEntry](ctx, CHEId)
		if err != nil {
			return nilVal, err
		}
	}

	switch CHEUI.CHEType {
	case "message":
		cheuis, err := cache.GetUser[UITypes.MsgSender](ctx, CHEUI.Sender.(string))
		if err != nil {
			return nilVal, err
		}

		cheuis.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(cheuis.ProfilePicUrl)

		CHEUI.Sender = cheuis

		CHEUI.Content = cloudStorageService.MessageMediaCloudNameToUrl(CHEUI.Content)

		userEmojiMap, err := cache.GetMsgReactions(ctx, CHEId)
		if err != nil {
			return nilVal, err
		}

		msgReactions := []UITypes.MsgReaction{}
		reactionsCount := make(map[string]int, 2)

		for user, emoji := range userEmojiMap {
			var msgr UITypes.MsgReaction

			msgr.Emoji = emoji
			msgr.Reactor, err = cache.GetUser[UITypes.MsgReactor](ctx, user)
			if err != nil {
				return nilVal, err
			}

			msgr.Reactor.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(msgr.Reactor.ProfilePicUrl)

			msgReactions = append(msgReactions, msgr)
			reactionsCount[emoji]++
		}

		CHEUI.Reactions = msgReactions
		CHEUI.ReactionsCount = reactionsCount
	}

	return CHEUI, nil
}
