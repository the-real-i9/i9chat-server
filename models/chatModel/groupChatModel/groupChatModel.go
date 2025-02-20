package groupChat

import (
	"context"
	"i9chat/helpers"
	"i9chat/models/db"
	"log"
	"maps"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type NewGroupChat struct {
	ClientData     map[string]any `json:"client_resp"`
	InitMemberData map[string]any `json:"init_member_resp"`
}

func New(ctx context.Context, clientUsername, name, description, pictureUrl string, initUsers []string, createdAt time.Time) (NewGroupChat, error) {
	var newGroupChat NewGroupChat

	res, err := db.Query(
		ctx,
		`
		MATCH (initUser:User WHERE initUser.username IN $init_users), (clientUser:User{ username: $client_username })
		CREATE (group:Group{ id: randomUUID(), name: $name, description: $description, picture_url: $picture_url, created_at: $created_at })

		WITH group, initUser, clientUser
		CREATE (clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
			(clientUser)-[:HAS_CHAT]->(clientChat:GroupChat{ owner_username: $client_username, group_id: $group.id, last_activity_type: "group activity", last_group_activity_at: $created_at, updated_at: $created_at })-[:WITH_GROUP]->(group),
			(clientUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: "You created " + $name, created_at: $created_at })-[:IN_GROUP_CHAT]->(clientChat),
			(clientUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: "You added " + $init_users_str, created_at: $created_at })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, initUser
		CREATE (initUser)-[:IS_MEMBER_OF { role: "member" }]->(group),
			(initUser)-[:HAS_CHAT]->(initUserChat:GroupChat{ owner_username: initUser.username, group_id: group.id, last_activity_type: "group activity", last_group_activity_at: $created_at, updated_at: $created_at })-[:WITH_GROUP]->(group),
			(initUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: $client_username + " created " + $name, created_at: $created_at })-[:IN_GROUP_CHAT]->(initUserChat),
			(initUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: "You were added", created_at: $created_at })-[:IN_GROUP_CHAT]->(initUserChat)

		WITH group
		RETURN group { .id, .name, .description, .picture_url, last_activity: { type: "group activity", info: "You added " + $init_users_str } } AS client_resp,
			group { .id, .name, .picture_url, last_activity: { type: "group activity", info: "You were added" } } AS init_member_resp
		`,
		map[string]any{
			"client_username": clientUsername,
			"name":            name,
			"description":     description,
			"picture_url":     pictureUrl,
			"init_users":      initUsers,
			"init_users_str":  strings.Join(initUsers, ", "),
			"created_at":      createdAt,
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: New:", err)
		return newGroupChat, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return newGroupChat, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying valid usernames in 'initUsers'")
	}

	helpers.MapToStruct(res.Records[0].AsMap(), &newGroupChat)

	return newGroupChat, nil
}

type NewActivity struct {
	ClientData      any      `json:"client_resp"`
	MemberData      any      `json:"member_resp"`
	MemberUsernames []string `json:"members_usernames"`
}

func ChangeName(ctx context.Context, groupId, clientUsername, newName string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
				
			SET clientChat.last_activity_type = "group activity",
				clientChat.last_group_activity_at = $at
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You changed group name from " + group.name + " to " + $new_name, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, cligact, group.name AS old_name

			SET group.name = $new_name

			WITH group, cligact, old_name

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username)

			RETURN cligact.info AS client_resp, collect(memberUser.username) AS member_usernames, old_name
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"new_name":        newName,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'groupId' | you're a member and an admin of this group")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		memberUsernames := resMap["member_usernames"].([]string)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "group activity",
					memberChat.last_group_activity_at = $at
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " changed group name from " + $old_name + " to " + $new_name, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

				RETURN memgact.info AS member_resp
				`,
				map[string]any{
					"group_id":         groupId,
					"client_username":  clientUsername,
					"new_name":         newName,
					"old_name":         resMap["old_name"],
					"member_usernames": memberUsernames,
					"at":               at,
				},
			)
			if err != nil {
				return nil, err
			}

			maps.Copy(res.Record().AsMap(), resMap)
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return newActivity, fe
		}

		log.Println("groupChatModel.go: ChangeName:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.AnyToStruct(res, &newActivity)

	return newActivity, nil
}

func ChangeDescription(ctx context.Context, groupId, clientUsername, newDescription string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
			SET clientChat.last_activity_type = "group activity",
				clientChat.last_group_activity_at = $at
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You changed group description from " + group.description + " to " + $new_description, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, cligact, group.description AS old_description

			SET group.description = $new_description

			WITH group, cligact, old_description

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username)

			RETURN cligact.info AS client_resp, collect(memberUser.username) AS member_usernames, old_description
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"new_description": newDescription,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'groupId' | you're a member and an admin of this group")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		memberUsernames := resMap["member_usernames"].([]string)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "group activity",
					memberChat.last_group_activity_at = $at
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " changed group description from " + $old_description + " to " + $new_description, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

				RETURN memgact.info AS member_resp
				`,
				map[string]any{
					"group_id":         groupId,
					"client_username":  clientUsername,
					"new_description":  newDescription,
					"old_description":  resMap["old_description"],
					"member_usernames": memberUsernames,
					"at":               at,
				},
			)
			if err != nil {
				return nil, err
			}

			maps.Copy(res.Record().AsMap(), resMap)
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return newActivity, fe
		}

		log.Println("groupChatModel.go: ChangeDescription:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.AnyToStruct(res, &newActivity)

	return newActivity, nil
}

func ChangePicture(ctx context.Context, groupId, clientUsername, newPictureUrl string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 3)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
			SET clientChat.last_activity_type = "group activity",
				clientChat.last_group_activity_at = $at
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You changed group picture", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			SET group.picture_url = $new_pic_url

			WITH group, cligact

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username)

			RETURN cligact.info AS client_resp, collect(memberUser.username) AS member_usernames
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"new_pic_url":     newPictureUrl,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'groupId' | you're a member and an admin of this group")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		memberUsernames := resMap["member_usernames"].([]string)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "group activity",
					memberChat.last_group_activity_at = $at
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " changed group picture", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

				RETURN memgact.info AS member_resp
				`,
				map[string]any{
					"group_id":         groupId,
					"client_username":  clientUsername,
					"member_usernames": memberUsernames,
					"at":               at,
				},
			)
			if err != nil {
				return nil, err
			}

			maps.Copy(res.Record().AsMap(), resMap)
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return newActivity, fe
		}

		log.Println("groupChatModel.go: ChangePicture:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.AnyToStruct(res, &newActivity)

	return newActivity, nil
}

func AddUsers(ctx context.Context, groupId, clientUsername string, newUsers []string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
				(newUser:User WHERE newUser.username IN $new_users AND NOT EXISTS { (newUser)-[:LEFT_GROUP]->(group) }
					AND NOT EXISTS { (newUser)-[:IS_MEMBER_OF]->(group) }

			SET clientChat.last_activity_type = "group activity",
				clientChat.last_group_activity_at = $at
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You added " + $new_users_str, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, newUser, cligact
			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username),
				(group)-[rur:REMOVED_USER]->(newUser) 
			DELETE rur

			WITH group, newUser, cligact, collect(memberUser.username) AS member_usernames
			CREATE (newUser)-[:IS_MEMBER_OF { role: "member" }]->(group)
			MERGE (newUser)-[:HAS_CHAT]->(newUserChat:GroupChat{ owner_username: newUser.username, group_id: $group_id })-[:WITH_GROUP]->(group)
			ON CREATE 
				SET newUserChat.updated_at = $at
			SET newUserChat.last_activity_type = "group activity",
				newUserChat.last_group_activity_at = $at
			CREATE (newUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: "You were added", created_at: $at })-[:IN_GROUP_CHAT]->(newUserChat)

			RETURN cligact.info AS client_resp, 
				group { .id, .name, .picture_url, last_activity: { type: "group activity", info: "You were added" } } AS new_user_resp,
				member_usernames
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"new_users":       newUsers,
				"new_users_str":   strings.Join(newUsers, ", "),
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a correct 'groupId' | valid usernames are in 'newUsers' | a 'newUser' isn't already a member | a 'newUser' hasn't already left this group | you're a member and an admin of this group")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		memberUsernames := resMap["member_usernames"].([]string)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "group activity",
					memberChat.last_group_activity_at = $at
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " added " + $new_users_str, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

				RETURN memgact.info AS member_resp
				`,
				map[string]any{
					"group_id":         groupId,
					"client_username":  clientUsername,
					"new_users_str":    strings.Join(newUsers, ", "),
					"member_usernames": memberUsernames,
					"at":               at,
				},
			)
			if err != nil {
				return nil, err
			}

			maps.Copy(res.Record().AsMap(), resMap)
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return newActivity, nil, fe
		}

		log.Println("groupChatModel.go: AddUsers:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	resMap := res.(map[string]any)

	helpers.AnyToStruct(res, &newActivity)

	return newActivity, resMap["new_user_resp"], nil
}

func RemoveUser(ctx context.Context, groupId, clientUsername, targetUser string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
				(group)<-[mem:IS_MEMBER_OF]-(targetUser:User{ username: $target_user })

			DELETE mem

			SET clientChat.last_activity_type = "group activity",
				clientChat.last_group_activity_at = $at
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You removed " + $target_user, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, targetUser, cligact
			MATCH (targetUserChat:GroupChat{ owner_username: targetUser.username, group_id: $group_id })

			SET targetUserChat.last_activity_type = "group activity",
				targetUserChat.last_group_activity_at = $at
			CREATE (group)-[:REMOVED_USER]->(targetUser),
				(targetUser)-[:RECEIVES_ACTIVITY]->(tugact:GroupActivity{ info: $client_username + " removed you", created_at: $at })-[:IN_GROUP_CHAT]-(targetUserChat)

			WITH group, cligact, tugact

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username)

			RETURN cligact.info AS client_resp, 
				tugact.info AS target_user_resp,
				collect(memberUser.username) AS member_usernames
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"target_user":     targetUser,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a correct 'groupId' | the username of 'user' is valid | 'user' is a member of this group | you're a member and an admin of this group")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		memberUsernames := resMap["member_usernames"].([]string)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "group activity",
					memberChat.last_group_activity_at = $at
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " removed " + $target_user, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

				RETURN memgact.info AS member_resp
				`,
				map[string]any{
					"group_id":         groupId,
					"client_username":  clientUsername,
					"target_user":      targetUser,
					"member_usernames": memberUsernames,
					"at":               at,
				},
			)
			if err != nil {
				return nil, err
			}

			maps.Copy(res.Record().AsMap(), resMap)
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return newActivity, nil, fe
		}

		log.Println("groupChatModel.go: RemoveUser:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	resMap := res.(map[string]any)

	helpers.AnyToStruct(res, &newActivity)

	return newActivity, resMap["target_user_resp"], nil
}

func Join(ctx context.Context, groupId, clientUsername string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 3)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (clientUser:User{ username: $client_username }), (group:Group{ id: $group_id })
			WHERE NOT EXISTS { (clientUser)-[:IS_MEMBER_OF]->(group) }
				AND NOT EXISTS { (group)-[:REMOVED_USER]->(clientUser) }

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User),
				(clientUser)-[lgr:LEFT_GROUP]->(group) 
			DELETE lgr

			WITH group, clientUser, collect(memberUser.username) AS member_usernames
			CREATE (clientUser)-[:IS_MEMBER_OF { role: "member" }]->(group)
			MERGE (clientUser)-[:HAS_CHAT]->(clientChat:GroupChat{ owner_username: clientUser.username, group_id: $group.id })-[:WITH_GROUP]->(group)
			ON CREATE 
				SET clientChat.updated_at = $at
			SET clientChat.last_activity_type = "group activity",
				clientChat.last_group_activity_at = $at
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: "You joined", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			RETURN group { .id, .name, .picture_url, last_activity: { type: "group activity", info: "You joined" } } AS client_resp,
				member_usernames
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a correct 'groupId' | you're not already a member of this group | you've not been previously removed from this group")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		memberUsernames := resMap["member_usernames"].([]string)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "group activity",
					memberChat.last_group_activity_at = $at
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " joined", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

				RETURN memgact.info AS member_resp
				`,
				map[string]any{
					"group_id":         groupId,
					"client_username":  clientUsername,
					"member_usernames": memberUsernames,
					"at":               at,
				},
			)
			if err != nil {
				return nil, err
			}

			maps.Copy(res.Record().AsMap(), resMap)
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return newActivity, fe
		}

		log.Println("groupChatModel.go: Join:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.AnyToStruct(res, &newActivity)

	return newActivity, nil
}

func Leave(ctx context.Context, groupId, clientUsername string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 3)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (group:Group{ id: $group_id })<-[mem:IS_MEMBER_OF]-(clientUser:User{ username: $client_username }),
				(clientChat:GroupChat{ owner_username: clientUser.username, group_id: $group_id })

			DELETE mem

			WITH group, clientUser, clientChat
			SET clientChat.last_activity_type = "group activity",
				clientChat.last_group_activity_at = $at
			CREATE (clientUser)-[:LEFT_GROUP]->(group),
				(clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You left", created_at: $at })-[:IN_GROUP_CHAT]-(clientChat)

			WITH group, cligact

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser)

			RETURN cligact.info AS client_resp,
				collect(memberUser.username) AS member_usernames
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a correct 'groupId' | you're a member of this group")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		memberUsernames := resMap["member_usernames"].([]string)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "group activity",
					memberChat.last_group_activity_at = $at
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " left", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

				RETURN memgact.info AS member_resp
				`,
				map[string]any{
					"group_id":         groupId,
					"client_username":  clientUsername,
					"member_usernames": memberUsernames,
					"at":               at,
				},
			)
			if err != nil {
				return nil, err
			}

			maps.Copy(res.Record().AsMap(), resMap)
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return newActivity, fe
		}

		log.Println("groupChatModel.go: Leave:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.AnyToStruct(res, &newActivity)

	return newActivity, nil
}

func MakeUserAdmin(ctx context.Context, groupId, clientUsername, targetUser string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
				(group)<-[mem:IS_MEMBER_OF { role: "member" }]-(targetUser:User{ username: $target_user })

			SET mem.role = "admin",
				clientChat.last_activity_type = "group activity",
				clientChat.last_group_activity_at = $at
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You made " + $target_user + " group admin", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, targetUser, cligact
			MATCH (targetUserChat:GroupChat{ owner_username: targetUser.username, group_id: $group_id })

			SET targetUserChat.last_activity_type = "group activity",
				targetUserChat.last_group_activity_at = $at
			CREATE (targetUser)-[:RECEIVES_ACTIVITY]->(tugact:GroupActivity{ info: $client_username + " made you group admin", created_at: $at })-[:IN_GROUP_CHAT]-(targetUserChat)

			WITH group, cligact, tugact
			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username NOT IN [$client_username, $target_user])

			RETURN cligact.info AS client_resp, 
				tugact.info AS target_user_resp,
				collect(memberUser.username) AS member_usernames
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"target_user":     targetUser,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a correct 'groupId' | 'user' is a member, and not already an admin of this group | you're a member and an admin this group")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		memberUsernames := resMap["member_usernames"].([]string)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "group activity",
					memberChat.last_group_activity_at = $at
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " made " + $target_user + " group admin", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

				RETURN memgact.info AS member_resp
				`,
				map[string]any{
					"group_id":         groupId,
					"client_username":  clientUsername,
					"target_user":      targetUser,
					"member_usernames": memberUsernames,
					"at":               at,
				},
			)
			if err != nil {
				return nil, err
			}

			maps.Copy(res.Record().AsMap(), resMap)
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return newActivity, nil, fe
		}

		log.Println("groupChatModel.go: MakeUserAdmin:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	resMap := res.(map[string]any)

	helpers.AnyToStruct(res, &newActivity)

	return newActivity, resMap["target_user_resp"], nil
}

func RemoveUserFromAdmins(ctx context.Context, groupId, clientUsername, targetUser string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
				(group)<-[mem:IS_MEMBER_OF { role: "admin" }]-(targetUser:User{ username: $target_user })

			SET mem.role = "member",
				clientChat.last_activity_type = "group activity",
				clientChat.last_group_activity_at = $at
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You removed " + $target_user + " from group admins", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, targetUser, cligact
			MATCH (targetUserChat:GroupChat{ owner_username: targetUser.username, group_id: $group_id })
			SET targetUserChat.last_activity_type = "group activity",
				targetUserChat.last_group_activity_at = $at
			CREATE (targetUser)-[:RECEIVES_ACTIVITY]->(tugact:GroupActivity{ info: $client_username + " removed you from group admins", created_at: $at })-[:IN_GROUP_CHAT]-(targetUserChat)

			WITH group, cligact, tugact
			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username NOT IN [$client_username, $target_user])

			RETURN cligact.info AS client_resp, 
				tugact.info AS target_user_resp,
				collect(memberUser.username) AS member_usernames
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"target_user":     targetUser,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a correct 'groupId' | 'user' is a member, and an admin of this group | you're a member and an admin this group")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		memberUsernames := resMap["member_usernames"].([]string)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "group activity",
					memberChat.last_group_activity_at = $at
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " removed " + $target_user + " from group admins", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

				RETURN memgact.info AS member_resp
				`,
				map[string]any{
					"group_id":         groupId,
					"client_username":  clientUsername,
					"target_user":      targetUser,
					"member_usernames": memberUsernames,
					"at":               at,
				},
			)
			if err != nil {
				return nil, err
			}

			maps.Copy(res.Record().AsMap(), resMap)
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return newActivity, nil, fe
		}

		log.Println("groupChatModel.go: RemoveUserFromAdmins:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	resMap := res.(map[string]any)

	helpers.AnyToStruct(res, &newActivity)

	return newActivity, resMap["target_user_resp"], nil
}

type NewMessage struct {
	ClientData      map[string]any `json:"client_resp"`
	MemberData      map[string]any `json:"member_resp"`
	MemberUsernames []string       `json:"members_usernames"`
}

func SendMessage(ctx context.Context, groupId, clientUsername string, msgContent []byte, createdAt time.Time) (NewMessage, error) {
	var newMessage NewMessage

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser)

			CREATE (message:GroupMessage{ id: randomUUID(), content: $message_content, delivery_status: "sent", created_at: $created_at })

			WITH group, clientChat, clientUser, message

			SET clientChat.last_activity_type = "message", 
				clientChat.updated_at = $created_at,
				clientChat.last_message_id = message.id
			CREATE (clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_GROUP_CHAT]->(clientChat)
			
			WITH group, message.id AS msgId

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username)

			RETURN { new_msg_id: msgId } AS client_resp,
				collect(memberUser.username) AS member_usernames, msgId AS new_msg_id
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'groupId' | you have a chat with this group (not necessarily a member)")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		memberUsernames := resMap["member_usernames"].([]string)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (message:GroupMessage{ id: $new_msg_id })<-[:SENDS_MESSAGE]-(clientUser)

				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "message", 
					memberChat.updated_at = $created_at,
					memberChat.last_message_id = message.id
				CREATE (memberUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_GROUP_CHAT]->(memberChat)

				WITH message, toString(message.created_at), clientUser { .username, .profile_pic_url } AS sender

				RETURN message { .*, created_at, sender } AS member_resp
				`,
				map[string]any{
					"group_id":         groupId,
					"client_username":  clientUsername,
					"member_usernames": memberUsernames,
					"new_msg_id":       resMap["new_msg_id"],
					"at":               at,
				},
			)
			if err != nil {
				return nil, err
			}

			maps.Copy(res.Record().AsMap(), resMap)
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return newMessage, fe
		}

		log.Println("groupChatModel.go: SendMessage:", err)
		return newMessage, fiber.ErrInternalServerError
	}

	helpers.AnyToStruct(res, &newMessage)

	return newMessage, nil
}

type MsgAck struct {
	All             bool     `json:"all"`
	MemberUsernames []string `json:"member_usernames"`
}

func AckMessageDelivered(ctx context.Context, clientUsername, groupId, msgId string, deliveredAt time.Time) (MsgAck, error) {
	var msgAck MsgAck

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 3)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientChat)<-[:IN_GROUP_CHAT]-(message:GroupMessage{ id: $message_id, delivery_status: "sent" })<-[:RECEIVES_MESSAGE]-()

			SET clientChat.unread_messages_count = coalesce(clientChat.unread_messages_count, 0) + 1
			CREATE (message)-[:DELIVERED_TO { at: $delivered_at }]->(clientUser)

			WITH group, message
			OPTIONAL MATCH (message)-[:DELIVERED_TO]->(delUser),
				(group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)

			RETURN collect(delUser.username) AS delv_to_usernames,
				collect(memberUser.username) AS member_usernames
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"message_id":      msgId,
				"delivered_at":    at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'groupId', and a valid 'msgId' received in the group chat | the message ('msgId') has not previously been acknowledged as 'delivered' or 'read'")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		// checking if the message has delivered to all members
		memberUsernames := resMap["member_usernames"].([]string)
		delvtoUsernames := resMap["delv_to_usernames"].([]string)

		delvToAll := helpers.AllAinB(memberUsernames, delvtoUsernames)

		resMap["all"] = delvToAll

		if delvToAll {
			_, err = tx.Run(
				ctx,
				`
				MATCH (message:GroupMessage{ id: $msg_id })
				SET message.delivery_status = "delivered"
				`,
				map[string]any{
					"msg_id": msgId,
				},
			)
			if err != nil {
				return nil, err
			}
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return msgAck, fe
		}

		log.Println("groupChatModel.go: AckMessageDelivered:", err)
		return msgAck, fiber.ErrInternalServerError
	}

	helpers.AnyToStruct(res, &msgAck)

	return msgAck, nil
}

func AckMessageRead(ctx context.Context, clientUsername, groupId, msgId string, readAt time.Time) (MsgAck, error) {
	var msgAck MsgAck

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 3)

		var (
			res neo4j.ResultWithContext
			err error
			at  = time.Now().UTC()
		)

		res, err = tx.Run(
			ctx,
			`
			MATCH (clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientChat)<-[:IN_GROUP_CHAT]-(message:GroupMessage{ id: $message_id } WHERE message.delivery_status <> "read")<-[:RECEIVES_MESSAGE]-()

			WITH clientChat, message, CASE coalesce(clientChat.unread_messages_count, 0) WHEN <> 0 THEN clientChat.unread_messages_count - 1 ELSE 0 END AS unread_messages_count
			SET clientChat.unread_messages_count = unread_messages_count
			CREATE (message)-[:READ_BY { at: $read_at } ]->(clientUser)

			WITH group, message
			OPTIONAL MATCH (message)-[:READ_BY]->(readUser),
				(group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)

			RETURN collect(readUser.username) AS read_by_usernames,
				collect(memberUser.username) AS member_usernames
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"message_id":      msgId,
				"delivered_at":    at,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'groupId', and a valid 'msgId' received in the group chat | the message ('msgId') has not previously been acknowledged as 'read'")
		}

		maps.Copy(res.Record().AsMap(), resMap)

		// checking if the message has delivered to all members
		memberUsernames := resMap["member_usernames"].([]string)
		readbyUsernames := resMap["read_by_usernames"].([]string)

		readByAll := helpers.AllAinB(memberUsernames, readbyUsernames)

		resMap["all"] = readByAll

		if readByAll {
			_, err = tx.Run(
				ctx,
				`
				MATCH (message:GroupMessage{ id: $msg_id })
				SET message.delivery_status = "read"
				`,
				map[string]any{
					"msg_id": msgId,
				},
			)
			if err != nil {
				return nil, err
			}
		}

		return resMap, nil
	})
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return msgAck, fe
		}

		log.Println("groupChatModel.go: AckMessageRead:", err)
		return msgAck, fiber.ErrInternalServerError
	}

	helpers.AnyToStruct(res, &msgAck)

	return msgAck, nil
}

func ReactToMessage(ctx context.Context, groupChatId, msgId, clientUsername string, reaction rune) error {
	return nil
}

type HistoryItem struct {
	HistoryItemType string `json:"hist_item_type"`

	// for message
	Id             string         `json:"id,omitempty"`
	Content        string         `json:"content,omitempty"`
	DeliveryStatus string         `json:"delivery_status,omitempty"`
	CreatedAt      string         `json:"created_at,omitempty"`
	Sender         map[string]any `json:"sender,omitempty"`

	// for group activity
	Info string `json:"info,omitempty"`
}

func GetChatHistory(ctx context.Context, clientUsername, groupId string, limit int, offset time.Time) ([]HistoryItem, error) {
	var chatHistory []HistoryItem

	res, err := db.Query(
		ctx,
		`
		MATCH (clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })
		CALL (clientChat) {
			OPTIONAL MATCH (clientChat)<-[:IN_GROUP_CHAT]-(message:GroupMessage),
				(message)<-[:SENDS_MESSAGE]-(senderUser),
				(message)<-[rxn:REACTS_TO_MESSAGE]-(reactorUser)
			WHERE message.created_at >= $offset
			WITH message, toString(message.created_at) AS created_at, senderUser { .username, .profile_pic_url } AS sender, collect({ user: reactorUser { .username, .profile_pic_url }, reaction: rxn.reaction }) AS reactions
			RETURN message { .*, created_at, sender, reactions, hist_item_type: "message" } AS hist_item, message.created_at AS created_at
		UNION
			MATCH (clientChat)<-[:IN_GROUP_CHAT]-(gactiv:GroupActivity)
			WHERE gactiv.created_at >= $offset
			RETURN { info: gactiv.info, hist_item_type: "activity" } AS hist_item, gactiv.created_at AS created_at
		}
		WITH hist_item, created_at ORDER BY created_at DESC
		LIMIT $limit

		RETURN collect(hist_item) AS chat_history
		`,
		map[string]any{
			"group_id":        groupId,
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: GetChatHistory", err)
		return chatHistory, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return chatHistory, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a correct 'groupId'")
	}

	ch, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "chat_history")

	helpers.AnyToStruct(ch, &chatHistory)

	return chatHistory, nil
}

func GetGroupInfo(ctx context.Context, groupId string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (group:Group{ id: $group_id }),
		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser),
			(group)<-[:IS_MEMBER_OF]-(memberUserOnline:User{ presence: "online" })

		WITH count(memberUser) AS members_count, count(memberUserOnline) AS online_members
		RETURN group { .name, .description, .picture_url, members_count, online_members } AS group_info
		`,
		map[string]any{
			"group_id": groupId,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: GetGroupInfo", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a correct 'groupId'")
	}

	gi, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "group_info")

	return gi, nil
}

func GetGroupMemInfo(ctx context.Context, clientUsername, groupId string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (group:Group{ id: $group_id }),
		OPTIONAL MATCH (group)<-[mr:IS_MEMBER_OF]-(isMember:User{ id: $client_username }),
			(group)-[:REMOVED_USER]->(userRemoved:User{ id: $client_username }),
			(group)<-[:LEFT_GROUP]-(userLeft:User{ id: $client_username })

		WITH CASE isMember WHEN IS NULL false ELSE true END AS member,
			CASE isMember WHEN IS NULL null ELSE mr.role END AS role,
			CASE userRemoved WHEN IS NULL false ELSE true END AS removed,
			CASE userLeft WHEN IS NULL false ELSE true END AS left
		RETURN { member, role, removed, left } AS group_mem_info
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: GetGroupMemInfo", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a correct 'groupId'")
	}

	gmi, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "group_mem_info")

	return gmi, nil
}
