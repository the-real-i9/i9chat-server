package modelHelpers

import (
	"context"
	"fmt"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/cache"
)

func BuildUserSnippetUIFromCache(ctx context.Context, username string) (userSnippetUI UITypes.UserSnippet, err error) {
	nilVal := UITypes.UserSnippet{}

	userSnippetUI, err = cache.GetUser[UITypes.UserSnippet](ctx, username)
	if err != nil {
		return nilVal, err
	}

	return userSnippetUI, nil
}

func BuildUserProfileUIFromCache(ctx context.Context, username string) (userProfileUI UITypes.UserProfile, err error) {
	nilVal := UITypes.UserProfile{}

	userProfileUI, err = cache.GetUser[UITypes.UserProfile](ctx, username)
	if err != nil {
		return nilVal, err
	}

	return userProfileUI, nil
}

func buildChatSnippetUIFromCache(ctx context.Context, clientUsername, chatIdent string) (chatSnippetUI UITypes.ChatSnippet, err error) {
	nilVal := UITypes.ChatSnippet{}

	chatSnippetUI, err = cache.GetChat[UITypes.ChatSnippet](ctx, clientUsername, chatIdent)
	if err != nil {
		return nilVal, err
	}

	switch chatSnippetUI.Type {
	case "direct":
		chatSnippetUI.PartnerUser, err = cache.GetUser[UITypes.ChatPartnerUser](ctx, chatSnippetUI.PartnerUser.(string))
		if err != nil {
			return nilVal, err
		}
	case "group":
		chatSnippetUI.Group, err = cache.GetGroup[UITypes.ChatGroup](ctx, chatSnippetUI.Group.(string))
		if err != nil {
			return nilVal, err
		}
	default:
		return nilVal, fmt.Errorf("unknown chatSnippetUI.Type:%s debug", chatSnippetUI.Type)
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
	default:
		return nilVal, fmt.Errorf("unknown chatType:%s debug", chatType)
	}

	switch CHEUI.CHEType {
	case "message":
		CHEUI.Sender, err = cache.GetUser[UITypes.MsgSender](ctx, CHEUI.Sender.(string))
		if err != nil {
			return nilVal, err
		}

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

			msgReactions = append(msgReactions, msgr)
			reactionsCount[emoji]++
		}

		CHEUI.Reactions = msgReactions
		CHEUI.ReactionsCount = reactionsCount

	default:
		return nilVal, fmt.Errorf("unknown CHEType:%s: debug", CHEUI.CHEType)
	}

	return CHEUI, nil
}
