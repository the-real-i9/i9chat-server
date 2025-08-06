package groupChat

import (
	"context"
	"fmt"
	"i9chat/src/helpers"
	"i9chat/src/models/db"
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
			(clientUser)-[:HAS_CHAT]->(clientChat:GroupChat{ owner_username: $client_username, group_id: group.id, last_activity_type: "group activity", last_group_activity_at: $created_at, updated_at: $created_at })-[:WITH_GROUP]->(group),
			(clientUser)-[:RECEIVES_ACTIVITY]->(:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You created " + $name, created_at: $created_at })-[:IN_GROUP_CHAT]->(clientChat),
			(clientUser)-[:RECEIVES_ACTIVITY]->(:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You added " + $init_users_str, created_at: $created_at })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, initUser
		CREATE (initUser)-[:IS_MEMBER_OF { role: "member" }]->(group),
			(initUser)-[:HAS_CHAT]->(initUserChat:GroupChat{ owner_username: initUser.username, group_id: group.id, last_activity_type: "group activity", last_group_activity_at: $created_at, updated_at: $created_at })-[:WITH_GROUP]->(group),
			(initUser)-[:RECEIVES_ACTIVITY]->(:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " created " + $name, created_at: $created_at })-[:IN_GROUP_CHAT]->(initUserChat),
			(initUser)-[:RECEIVES_ACTIVITY]->(:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You were added", created_at: $created_at })-[:IN_GROUP_CHAT]->(initUserChat)

		WITH group
		RETURN group { .id, .name, .description, .picture_url, history: [{ chat_hist_entry_type: "group activity", info: "You created " + $name, created_at: $created_at }, { chat_hist_entry_type: "group activity", info: "You added " + $init_users_str, created_at: $created_at }] } AS client_resp,
			group { .id, .name, .description, .picture_url, history: [{ chat_hist_entry_type: "group activity", info: $client_username + " created " + $name, created_at: $created_at }, { chat_hist_entry_type: "group activity", info: "You were added", created_at: $created_at }] } AS init_member_resp
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
		return newGroupChat, nil
	}

	helpers.ToStruct(res.Records[0].AsMap(), &newGroupChat)

	return newGroupChat, nil
}

type NewActivity struct {
	ClientData      any      `json:"client_resp"`
	MemberData      any      `json:"member_resp"`
	MemberUsernames []string `json:"member_usernames"`
}

func ChangeName(ctx context.Context, groupId, clientUsername, newName string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
				
			SET clientChat.last_activity_type = "group activity",
				clientChat.last_group_activity_at = $at
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You changed group name from " + group.name + " to " + $new_name, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, cligact, group.name AS old_name

			SET group.name = $new_name

			WITH group, cligact, old_name

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)

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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		memberUsernames := resMap["member_usernames"].([]any)

		if len(memberUsernames) > 0 {
			res, err := tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				SET memberChat.last_activity_type = "group activity",
					memberChat.last_group_activity_at = $at
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " changed group name from " + $old_name + " to " + $new_name, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

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

			if !res.Next(ctx) {
				return nil, fmt.Errorf("crosscheck possible logical error")
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("groupChatModel.go: ChangeName:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &newActivity)

	return newActivity, nil
}

func ChangeDescription(ctx context.Context, groupId, clientUsername, newDescription string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
			
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You changed group description from " + group.description + " to " + $new_description, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, cligact, group.description AS old_description

			SET group.description = $new_description

			WITH group, cligact, old_description

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)

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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		memberUsernames := resMap["member_usernames"].([]any)

		if len(memberUsernames) > 0 {
			res, err := tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " changed group description from " + $old_description + " to " + $new_description, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

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

			if !res.Next(ctx) {
				return nil, fmt.Errorf("crosscheck possible logical error")
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("groupChatModel.go: ChangeDescription:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &newActivity)

	return newActivity, nil
}

func ChangePicture(ctx context.Context, groupId, clientUsername, newPictureUrl string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 3)

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
			
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You changed group picture", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			SET group.picture_url = $new_pic_url

			WITH group, cligact

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)

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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		memberUsernames := resMap["member_usernames"].([]any)

		if len(memberUsernames) > 0 {
			res, err := tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " changed group picture", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

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

			if !res.Next(ctx) {
				return nil, fmt.Errorf("crosscheck possible logical error")
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("groupChatModel.go: ChangePicture:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &newActivity)

	return newActivity, nil
}

func AddUsers(ctx context.Context, groupId, clientUsername string, newUsers []string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
				(newUser:User WHERE newUser.username IN $new_users AND NOT EXISTS { (newUser)-[:LEFT_GROUP]->(group) }
					AND NOT EXISTS { (newUser)-[:IS_MEMBER_OF]->(group) })

			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You added " + $new_users_str, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, newUser, cligact
			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)
			OPTIONAL MATCH(group)-[rur:REMOVED_USER]->(newUser) 

			DELETE rur

			WITH group, newUser, cligact, collect(memberUser.username) AS member_usernames
			CREATE (newUser)-[:IS_MEMBER_OF { role: "member" }]->(group)
			MERGE (newUser)-[:HAS_CHAT]->(newUserChat:GroupChat{ owner_username: newUser.username, group_id: $group_id })-[:WITH_GROUP]->(group)
			
			CREATE (newUser)-[:RECEIVES_ACTIVITY]->(:GroupChatEntry{ chat_hist_entry_type: "group activity",  info: "You were added", created_at: $at })-[:IN_GROUP_CHAT]->(newUserChat)

			RETURN cligact.info AS client_resp, 
				group { .id, .name, .description, .picture_url, history: [{ chat_hist_entry_type: "group activity", info: "You were added", created_at: $at }] } AS new_user_resp,
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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		memberUsernames := resMap["member_usernames"].([]any)

		if len(memberUsernames) > 0 {
			res, err := tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " added " + $new_users_str, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

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

			if !res.Next(ctx) {
				return nil, fmt.Errorf("crosscheck possible logical error")
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("groupChatModel.go: AddUsers:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	resMap := res.(map[string]any)

	helpers.ToStruct(res, &newActivity)

	return newActivity, resMap["new_user_resp"], nil
}

func RemoveUser(ctx context.Context, groupId, clientUsername, targetUser string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
				(group)<-[mem:IS_MEMBER_OF]-(targetUser:User{ username: $target_user })

			DELETE mem

			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You removed " + $target_user, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, targetUser, cligact
			MATCH (targetUserChat:GroupChat{ owner_username: targetUser.username, group_id: $group_id })

			CREATE (group)-[:REMOVED_USER]->(targetUser),
				(targetUser)-[:RECEIVES_ACTIVITY]->(tugact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " removed you", created_at: $at })-[:IN_GROUP_CHAT]->(targetUserChat)

			WITH group, cligact, tugact

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)

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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		memberUsernames := resMap["member_usernames"].([]any)

		if len(memberUsernames) > 0 {
			res, err := tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " removed " + $target_user, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

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

			if !res.Next(ctx) {
				return nil, fmt.Errorf("crosscheck possible logical error")
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("groupChatModel.go: RemoveUser:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	resMap := res.(map[string]any)

	helpers.ToStruct(res, &newActivity)

	return newActivity, resMap["target_user_resp"], nil
}

func Join(ctx context.Context, groupId, clientUsername string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 3)

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (clientUser:User{ username: $client_username }), (group:Group{ id: $group_id })
			WHERE NOT EXISTS { (clientUser)-[:IS_MEMBER_OF]->(group) }
				AND NOT EXISTS { (group)-[:REMOVED_USER]->(clientUser) }

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser)
			OPTIONAL MATCH (clientUser)-[lgr:LEFT_GROUP]->(group)

			DELETE lgr

			WITH group, clientUser, collect(memberUser.username) AS member_usernames
			CREATE (clientUser)-[:IS_MEMBER_OF { role: "member" }]->(group)
			MERGE (clientUser)-[:HAS_CHAT]->(clientChat:GroupChat{ owner_username: clientUser.username, group_id: group.id })-[:WITH_GROUP]->(group)
			
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You joined", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			RETURN group { .id, .name, .description, .picture_url, history: [{ chat_hist_entry_type: "group activity", info: "You joined", created_at: $at }] } AS client_resp,
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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		memberUsernames := resMap["member_usernames"].([]any)

		if len(memberUsernames) > 0 {
			res, err := tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " joined", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

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

			if !res.Next(ctx) {
				return nil, nil
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("groupChatModel.go: Join:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &newActivity)

	return newActivity, nil
}

func Leave(ctx context.Context, groupId, clientUsername string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 3)

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (group:Group{ id: $group_id })<-[mem:IS_MEMBER_OF]-(clientUser:User{ username: $client_username }),
				(clientChat:GroupChat{ owner_username: clientUser.username, group_id: $group_id })

			DELETE mem

			WITH group, clientUser, clientChat
			CREATE (clientUser)-[:LEFT_GROUP]->(group),
				(clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You left", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		memberUsernames := resMap["member_usernames"].([]any)

		if len(memberUsernames) > 0 {
			res, err := tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " left", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

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

			if !res.Next(ctx) {
				return nil, fmt.Errorf("crosscheck possible logical error")
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("groupChatModel.go: Leave:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &newActivity)

	return newActivity, nil
}

func MakeUserAdmin(ctx context.Context, groupId, clientUsername, targetUser string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
				(group)<-[mem:IS_MEMBER_OF { role: "member" }]-(targetUser:User{ username: $target_user })

			SET mem.role = "admin"

			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You made " + $target_user + " group admin", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, targetUser, cligact
			MATCH (targetUserChat:GroupChat{ owner_username: targetUser.username, group_id: $group_id })

			CREATE (targetUser)-[:RECEIVES_ACTIVITY]->(tugact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " made you group admin", created_at: $at })-[:IN_GROUP_CHAT]->(targetUserChat)

			WITH group, cligact, tugact
			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE NOT memberUser.username IN [$client_username, $target_user])

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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		memberUsernames := resMap["member_usernames"].([]any)

		if len(memberUsernames) > 0 {
			res, err := tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " made " + $target_user + " group admin", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

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

			if !res.Next(ctx) {
				return nil, nil
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("groupChatModel.go: MakeUserAdmin:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	resMap := res.(map[string]any)

	helpers.ToStruct(res, &newActivity)

	return newActivity, resMap["target_user_resp"], nil
}

func RemoveUserFromAdmins(ctx context.Context, groupId, clientUsername, targetUser string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
				(group)<-[mem:IS_MEMBER_OF { role: "admin" }]-(targetUser:User{ username: $target_user })

			SET mem.role = "member"
				
			CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: "You removed " + $target_user + " from group admins", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, targetUser, cligact
			MATCH (targetUserChat:GroupChat{ owner_username: targetUser.username, group_id: $group_id })
			
			CREATE (targetUser)-[:RECEIVES_ACTIVITY]->(tugact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " removed you from group admins", created_at: $at })-[:IN_GROUP_CHAT]->(targetUserChat)

			WITH group, cligact, tugact
			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE NOT memberUser.username IN [$client_username, $target_user])

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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		memberUsernames := resMap["member_usernames"].([]any)

		if len(memberUsernames) > 0 {
			res, err = tx.Run(
				ctx,
				`
				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				
				CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupChatEntry{ chat_hist_entry_type: "group activity", info: $client_username + " removed " + $target_user + " from group admins", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

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

			if !res.Next(ctx) {
				return nil, fmt.Errorf("crosscheck possible logical error")
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("groupChatModel.go: RemoveUserFromAdmins:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	resMap := res.(map[string]any)

	helpers.ToStruct(res, &newActivity)

	return newActivity, resMap["target_user_resp"], nil
}

type NewMessage struct {
	ClientData      map[string]any `json:"client_resp"`
	MemberData      map[string]any `json:"member_resp"`
	MemberUsernames []string       `json:"member_usernames"`
}

func SendMessage(ctx context.Context, clientUsername, groupId, msgContent string, at time.Time) (NewMessage, error) {
	var newMessage NewMessage

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 4)

		res, err := tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser)
			WHERE EXISTS { (clientUser)-[:IS_MEMBER_OF]->(group) }

			CREATE (message:GroupMessage:GroupChatEntry{ id: randomUUID(), chat_hist_entry_type: "message", content: $message_content, delivery_status: "sent", created_at: $at })

			WITH group, clientChat, clientUser, message

			CREATE (clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_GROUP_CHAT]->(clientChat)
			
			WITH group, message.id AS msgId

			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)

			RETURN { new_msg_id: msgId } AS client_resp,
				collect(memberUser.username) AS member_usernames, msgId AS new_msg_id
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"message_content": msgContent,
				"at":              at,
			},
		)
		if err != nil {
			return nil, err
		}

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		memberUsernames := resMap["member_usernames"].([]any)

		if len(memberUsernames) > 0 {
			res, err := tx.Run(
				ctx,
				`
				MATCH (message:GroupMessage{ id: $new_msg_id })<-[:SENDS_MESSAGE]-(clientUser)

				MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username IN $member_usernames),
					(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
				
				CREATE (memberUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_GROUP_CHAT]->(memberChat)

				WITH message, toString(message.created_at) AS created_at, clientUser { .username, .profile_pic_url } AS sender

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

			if !res.Next(ctx) {
				return nil, fmt.Errorf("crosscheck possible logical error")
			}

			maps.Copy(resMap, res.Record().AsMap())
		}

		return resMap, nil
	})
	if err != nil {
		log.Println("groupChatModel.go: SendMessage:", err)
		return newMessage, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &newMessage)

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

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientChat)<-[:IN_GROUP_CHAT]-(message:GroupMessage{ id: $message_id, delivery_status: "sent" })<-[:RECEIVES_MESSAGE]-(clientUser),
				(msgSender)-[:SENDS_MESSAGE]->(message)

			SET clientChat.unread_messages_count = coalesce(clientChat.unread_messages_count, 0) + 1
			CREATE (message)-[:DELIVERED_TO { at: $delivered_at }]->(clientUser)

			WITH group, message, msgSender
			OPTIONAL MATCH (message)-[:DELIVERED_TO]->(delUser)
			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE NOT memberUser.username IN [msgSender.username, $client_username])

			RETURN collect(DISTINCT delUser.username) AS delv_to_usernames,
				collect(DISTINCT memberUser.username) AS member_usernames
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

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		// checking if the message has delivered to all members
		memberUsernames := resMap["member_usernames"].([]any)
		delvtoUsernames := resMap["delv_to_usernames"].([]any)

		delvToAll := helpers.AsubsetB(memberUsernames, delvtoUsernames)

		resMap["all"] = delvToAll

		if delvToAll {
			_, err := tx.Run(
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
		log.Println("groupChatModel.go: AckMessageDelivered:", err)
		return msgAck, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &msgAck)

	return msgAck, nil
}

func AckMessageRead(ctx context.Context, clientUsername, groupId, msgId string, readAt time.Time) (MsgAck, error) {
	var msgAck MsgAck

	res, err := db.MultiQuery(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		resMap := make(map[string]any, 3)

		at := time.Now().UTC()

		res, err := tx.Run(
			ctx,
			`
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientChat)<-[:IN_GROUP_CHAT]-(message:GroupMessage{ id: $message_id } WHERE message.delivery_status <> "read")<-[:RECEIVES_MESSAGE]-(clientUser),
				(msgSender)-[:SENDS_MESSAGE]->(message)

			WITH group, clientUser, clientChat, message, msgSender, CASE coalesce(clientChat.unread_messages_count, 0) WHEN <> 0 THEN clientChat.unread_messages_count - 1 ELSE 0 END AS unread_messages_count
			SET clientChat.unread_messages_count = unread_messages_count
			CREATE (message)-[:READ_BY { at: $read_at } ]->(clientUser)

			WITH group, message, msgSender
			OPTIONAL MATCH (message)-[:READ_BY]->(readUser)
			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE NOT memberUser.username IN [msgSender.username, $client_username])

			RETURN collect(DISTINCT readUser.username) AS read_by_usernames,
				collect(DISTINCT memberUser.username) AS member_usernames
			`,
			map[string]any{
				"client_username": clientUsername,
				"group_id":        groupId,
				"message_id":      msgId,
				"read_at":         at,
			},
		)
		if err != nil {
			return nil, err
		}

		if !res.Next(ctx) {
			return nil, nil
		}

		maps.Copy(resMap, res.Record().AsMap())

		// checking if the message has delivered to all members
		memberUsernames := resMap["member_usernames"].([]any)
		readbyUsernames := resMap["read_by_usernames"].([]any)

		readByAll := helpers.AsubsetB(memberUsernames, readbyUsernames)

		resMap["all"] = readByAll

		if readByAll {
			_, err := tx.Run(
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
		log.Println("groupChatModel.go: AckMessageRead:", err)
		return msgAck, fiber.ErrInternalServerError
	}

	helpers.ToStruct(res, &msgAck)

	return msgAck, nil
}

func ReactToMessage(ctx context.Context, groupChatId, msgId, clientUsername, reaction string, at time.Time) error {
	return nil
}

type ChatHistoryEntry struct {
	EntryType string `json:"chat_hist_entry_type"`
	CreatedAt string `json:"created_at"`

	// for group message
	Id             string         `json:"id,omitempty"`
	Content        string         `json:"content,omitempty"`
	DeliveryStatus string         `json:"delivery_status,omitempty"`
	Sender         map[string]any `json:"sender,omitempty"`

	// for group activity
	Info string `json:"info,omitempty"`
}

func ChatHistory(ctx context.Context, clientUsername, groupId string, limit int, offset time.Time) ([]ChatHistoryEntry, error) {
	var chatHistory []ChatHistoryEntry

	res, err := db.Query(
		ctx,
		`
		MATCH (clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })
		OPTIONAL MATCH (clientChat)<-[:IN_GROUP_CHAT]-(entry:GroupChatEntry WHERE entry.created_at < $offset)
		OPTIONAL MATCH (entry)<-[:SENDS_MESSAGE]-(senderUser)
		OPTIONAL MATCH (entry)<-[rxn:REACTS_TO_MESSAGE]-(reactorUser)
		
		WITH entry, 
			toString(entry.created_at) AS created_at, 
			senderUser { .username, .profile_pic_url } AS sender, 
			collect({ user: reactorUser { .username, .profile_pic_url }, reaction: rxn.reaction }) AS reactions
		ORDER BY created_at
		LIMIT $limit
		
		RETURN collect(entry { .*, created_at, sender, reactions }) AS chat_history
		`,
		map[string]any{
			"group_id":        groupId,
			"client_username": clientUsername,
			"limit":           limit,
			"offset":          offset,
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: GetChatHistory", err)
		return chatHistory, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return chatHistory, nil
	}

	ch, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "chat_history")

	helpers.ToStruct(ch, &chatHistory)

	return chatHistory, nil
}

func GroupInfo(ctx context.Context, groupId string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (group:Group{ id: $group_id })

		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser)
		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUserOnline:User{ presence: "online" })

		WITH group, count(DISTINCT memberUser) AS members_count, count(DISTINCT memberUserOnline) AS members_online_count
		RETURN group { .name, .description, .picture_url, members_count, members_online_count } AS group_info
		`,
		map[string]any{
			"group_id": groupId,
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: GroupInfo", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	gi, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "group_info")

	return gi, nil
}

func GroupMemInfo(ctx context.Context, clientUsername, groupId string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (group:Group{ id: $group_id }), (clientUser:User { username: $client_username })

		OPTIONAL MATCH (group)<-[member:IS_MEMBER_OF]-(clientUser)
		OPTIONAL MATCH (group)-[userRemoved:REMOVED_USER]->(clientUser)
		OPTIONAL MATCH (group)<-[userLeft:LEFT_GROUP]-(clientUser)

		WITH group,
			CASE member 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS is_member,
			CASE member 
				WHEN IS NULL THEN null 
				ELSE member.role 
			END AS user_role,
			CASE userRemoved 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS user_removed,
			CASE userLeft 
				WHEN IS NULL THEN false 
				ELSE true 
			END AS user_left

		RETURN group { is_member, user_role, user_removed, user_left } AS group_mem_info
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: GroupMemInfo", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	gmi, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "group_mem_info")

	return gmi, nil
}
