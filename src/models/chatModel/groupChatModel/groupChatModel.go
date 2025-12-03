package groupChat

import (
	"context"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/helpers"
	"i9chat/src/models/db"
	"i9chat/src/models/modelHelpers"

	"github.com/gofiber/fiber/v2"
)

type NewGroup struct {
	Id             string         `json:"id" db:"id"`
	Name           string         `json:"name" db:"name"`
	Description    string         `json:"description" db:"description"`
	PictureUrl     string         `json:"picture_url" db:"picture_url"`
	CreatedAt      int64          `json:"created_at" db:"created_at"`
	InitUsers      []any          `json:"-" db:"init_users"`
	ClientUserCHEs []any          `json:"-" db:"client_user_ches"`
	InitUsersCHEs  map[string]any `json:"-" db:"init_users_ches"`
}

func New(ctx context.Context, clientUsername, name, description, pictureUrl string, initUsers []string, createdAt int64) (NewGroup, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (initUser:User WHERE initUser.username IN $init_users), (clientUser:User{ username: $client_username })

		WITH collect(initUser) AS initUserRows, head(collect(clientUser)) AS clientUser

		CREATE (group:Group{ id: randomUUID(), name: $name, description: $description, picture_url: $picture_url, created_at: $created_at })

		CREATE (clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
			(clientUser)-[:HAS_CHAT]->(clientChat:GroupChat{ owner_username: $client_username, group_id: group.id })-[:WITH_GROUP]->(group),
			(cligact1:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You created " + $name })-[:IN_GROUP_CHAT]->(clientChat),
			(cligact2:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You added " + $init_users_str })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, initUserRows, [cligact1 {.che_id, .che_type, .info}, cligact2 {.che_id, .che_type, .info}] AS clientUserCHEs
		UNWIND initUserRows AS initUser

		WITH group, initUser, clientUserCHEs
		CREATE (initUser)-[:IS_MEMBER_OF { role: "member" }]->(group),
			(initUser)-[:HAS_CHAT]->(initUserChat:GroupChat{ owner_username: initUser.username, group_id: group.id })-[:WITH_GROUP]->(group),
			(initusergact1:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: $client_username + " created " + $name })-[:IN_GROUP_CHAT]->(initUserChat),
			(initusergact2:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You were added" })-[:IN_GROUP_CHAT]->(initUserChat)

		LET initUsersGactColl = collect({ inituser: initUser.username, gact1: initusergact1, gact2: initusergact2})
			initUsersCHEs = reduce(m = {}, x IN initUsersGactColl | m + [x.inituser]: [x.gact1 {.che_id, .che_type, .info}, x.gact2 {.che_id, .che_type, .info}])

		WITH group, clientUserCHEs, initUsersCHEs
		RETURN group { .id, .name, .description, .picture_url, .created_at, init_users: $init_users, creator_user_history: clientUserCHEs, init_users_history: initUsersCHEs } AS new_group
		`,
		map[string]any{
			"client_username": clientUsername,
			"name":            name,
			"description":     description,
			"picture_url":     pictureUrl,
			"init_users":      initUsers,
			"init_users_str":  helpers.JoinWithCommaAnd(initUsers...),
			"created_at":      createdAt,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return NewGroup{}, fiber.ErrInternalServerError
	}

	newGroup := modelHelpers.RKeyGet[NewGroup](res.Records, "new_group")

	return newGroup, nil
}

type EditActivity struct {
	ClientUserCHE   map[string]any `json:"-" db:"client_user_che"`
	MemberUsersCHE  map[string]any `json:"-" db:"member_users_che"`
	MemberUsernames []any          `json:"-" db:"member_usernames"`
}

func ChangeName(ctx context.Context, groupId, clientUsername, newName string) (EditActivity, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
			
		CREATE (cligact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You changed group name from " + group.name + " to " + $new_name })-[:IN_GROUP_CHAT]->(clientChat)

		LET old_name = group.name

		SET group.name = $new_name

		WITH cligact { .che_id, .che_type, .info } AS clientUserCHE, old_name, group

		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(group)

		LET memberUsernames = collect(memberUser.username),
			memberUsersCHE = reduce(m = {}, mu IN memberUsernames | m + [mu]: { che_id: randomUUID(), che_type: "group activity", info: $client_username + " changed group name from " + old_name + " to " + $new_name })

		FOREACH (mc IN collect(memberChat) | CREATE (:GroupChatEntry{ che_id: memberUsersCHE[mc.owner_username].che_id, che_type: memberUsersCHE[mc.owner_username].che_type, info: memberUsersCHE[mc.owner_username].info })-[:IN_GROUP_CHAT]->(mc))

		WITH DISTINCT clientUserCHE, memberUsersCHE, memberUsernames

		RETURN { client_user_che: clientUserCHE , member_users_che: memberUsersCHE, member_usernames: memberUsernames } AS new_group_activity
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"new_name":        newName,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return EditActivity{}, fiber.ErrInternalServerError
	}

	newGact := modelHelpers.RKeyGet[EditActivity](res.Records, "new_group_activity")

	return newGact, nil
}

func ChangeDescription(ctx context.Context, groupId, clientUsername, newDescription string) (EditActivity, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
		
		CREATE (cligact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You changed group description from " + group.description + " to " + $new_description })-[:IN_GROUP_CHAT]->(clientChat)

		LET old_description = group.description

		SET group.description = $new_description

		WITH cligact { .che_id, .che_type, .info } AS clientUserCHE, 
			old_description, group

		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(group)

		LET memberUsernames = collect(memberUser.username),
			memberUsersCHE = reduce(m = {}, mu IN memberUsernames | m + [mu]: { che_id: randomUUID(), che_type: "group activity", info: $client_username + " changed group description from " + old_description + " to " + $new_description })

		FOREACH (mc IN collect(memberChat) | CREATE (:GroupChatEntry{ che_id: memberUsersCHE[mc.owner_username].che_id, che_type: memberUsersCHE[mc.owner_username].che_type, info: memberUsersCHE[mc.owner_username].info })-[:IN_GROUP_CHAT]->(mc))

		WITH DISTINCT clientUserCHE, memberUsersCHE, memberUsernames

		RETURN { client_user_che: clientUserCHE , member_users_che: memberUsersCHE, member_usernames: memberUsernames } AS new_group_activity
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"new_description": newDescription,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return EditActivity{}, fiber.ErrInternalServerError
	}

	newGact := modelHelpers.RKeyGet[EditActivity](res.Records, "new_group_activity")

	return newGact, nil
}

func ChangePicture(ctx context.Context, groupId, clientUsername, newPictureUrl string) (EditActivity, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
		
		CREATE (cligact:GroupChatEntry{ che_type: "group activity", info: "You changed group picture" })-[:IN_GROUP_CHAT]->(clientChat)

		SET group.picture_url = $new_pic_url

		WITH cligact { .che_id, .che_type, .info } AS clientUserCHE, group

		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(group)

		LET memberUsernames = collect(memberUser.username),
			memberUsersCHE = reduce(m = {}, mu IN memberUsernames | m + [mu]: { che_id: randomUUID(), che_type: "group activity", info: $client_username + " changed group picture" })

		FOREACH (mc IN collect(memberChat) | CREATE (:GroupChatEntry{ che_id: memberUsersCHE[mc.owner_username].che_id, che_type: memberUsersCHE[mc.owner_username].che_type, info: memberUsersCHE[mc.owner_username].info })-[:IN_GROUP_CHAT]->(mc))

		WITH DISTINCT clientUserCHE, memberUsersCHE, memberUsernames

		RETURN { client_user_che: clientUserCHE , member_users_che: memberUsersCHE, member_usernames: memberUsernames } AS new_group_activity
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"new_pic_url":     newPictureUrl,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return EditActivity{}, fiber.ErrInternalServerError
	}

	newGact := modelHelpers.RKeyGet[EditActivity](res.Records, "new_group_activity")

	return newGact, nil
}

type AddUsersActivity struct {
	GroupInfo       map[string]any `json:"-" db:"group_info"`
	ClientUserCHE   map[string]any `json:"-" db:"client_user_che"`
	NewUsersCHE     map[string]any `json:"-" db:"new_users_che"`
	MemberUsersCHE  map[string]any `json:"-" db:"member_users_che"`
	NewUsernames    []any          `json:"-" db:"new_usernames"`
	MemberUsernames []any          `json:"-" db:"member_usernames"`
}

func AddUsers(ctx context.Context, groupId, clientUsername string, newUsers []string) (AddUsersActivity, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
			(newUser:User WHERE newUser.username IN $new_users AND NOT EXISTS { (newUser)-[:LEFT_GROUP]->(group) }
				AND NOT EXISTS { (newUser)-[:IS_MEMBER_OF]->(group) })
			
		WITH collect(newUser) AS nuRows
			head(collect(group)) AS group,
			head(collect(clientUser)) AS clientUser,
			head(collect(clientChat)) AS clientChat

		CREATE (cligact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You added " + $new_users_str })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, group AS canNullG, nuRows, cligact { .che_id, .che_type, .info } AS clientUserCHE
		OPTIONAL MATCH (canNullG)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(canNullG)

		LET memberUsernames = collect(memberUser.username),
			memberUsersCHE = reduce(m = {}, mu IN memberUsernames | m + [mu]: { che_id: randomUUID(), che_type: "group activity", info: $client_username + " added " + $new_users_str })

		FOREACH (mc IN collect(memberChat) | CREATE (:GroupChatEntry{ che_id: memberUsersCHE[mc.owner_username].che_id, che_type: memberUsersCHE[mc.owner_username].che_type, info: memberUsersCHE[mc.owner_username].info })-[:IN_GROUP_CHAT]->(mc))

		WITH group, nuRows, clientUserCHE, memberUsersCHE, memberUsernames
		UNWIND nuRows AS newUser

		LET canNullG = group, canNullNU = newUser

		OPTIONAL MATCH(canNullG)-[rur:REMOVED_USER]->(canNullNU)

		DELETE rur

		WITH group, newUser, nuRows, clientUserCHE, memberUsersCHE, memberUsernames
		CREATE (newUser)-[:IS_MEMBER_OF { role: "member" }]->(group)
		MERGE (newUser)-[:HAS_CHAT]->(newUserChat:GroupChat{ owner_username: newUser.username, group_id: $group_id })-[:WITH_GROUP]->(group)

		CREATE (nugact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity",  info: "You were added" })-[:IN_GROUP_CHAT]->(newUserChat)

		LET newUsernames = [nu IN nuRows | nu.username],
			newUsersGactColl = collect({ newuser: newUser.username, gact: nugact}),
			newUsersCHE = reduce(m = {}, x IN newUsersGactColl | m + [x.newuser]: x.gact {.che_id, .che_type, .info})

		WITH DISTINCT group { .id, .name, .description, .picture_url, .created_at } AS groupInfo, clientUserCHE, newUsersCHE, memberUsersCHE, newUsernames, memberUsernames

		RETURN { group_info: groupInfo, client_user_che: clientUserCHE, new_users_che: newUsersCHE, member_users_che: memberUsersCHE, new_usernames: newUsernames, member_usernames: memberUsernames } AS new_group_activity
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"new_users":       newUsers,
			"new_users_str":   helpers.JoinWithCommaAnd(newUsers...),
		},
	)
	if err != nil {
		helpers.LogError(err)
		return AddUsersActivity{}, fiber.ErrInternalServerError
	}

	newGact := modelHelpers.RKeyGet[AddUsersActivity](res.Records, "new_group_activity")

	return newGact, nil
}

type RemoveUserActivity struct {
	ClientUserCHE   map[string]any `json:"-" db:"client_user_che"`
	TargetUserCHE   map[string]any `json:"-" db:"target_user_che"`
	MemberUsersCHE  map[string]any `json:"-" db:"member_users_che"`
	MemberUsernames []any          `json:"-" db:"member_usernames"`
}

func RemoveUser(ctx context.Context, groupId, clientUsername, targetUser string) (RemoveUserActivity, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
			(group)<-[mem:IS_MEMBER_OF]-(targetUser:User{ username: $target_user })

		DELETE mem

		CREATE (cligact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You removed " + $target_user })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, targetUser, cligact { .che_id, .che_type, .info } AS clientUserCHE
		MATCH (targetUserChat:GroupChat{ owner_username: targetUser.username, group_id: $group_id })

		CREATE (group)-[:REMOVED_USER]->(targetUser),
			(tugact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: $client_username + " removed you" })-[:IN_GROUP_CHAT]->(targetUserChat)

		WITH clientUserCHE, tugact { .che_id, .che_type, .info } AS targetUserCHE, group

		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(group)

		LET memberUsernames = collect(memberUser.username),
			memberUsersCHE = reduce(m = {}, mu IN memberUsernames | m + [mu]: { che_id: randomUUID(), che_type: "group activity", info: $client_username + " removed " + $target_user })

		FOREACH (mc IN collect(memberChat) | CREATE (:GroupChatEntry{ che_id: memberUsersCHE[mc.owner_username].che_id, che_type: memberUsersCHE[mc.owner_username].che_type, info: memberUsersCHE[mc.owner_username].info })-[:IN_GROUP_CHAT]->(mc))

		WITH DISTINCT clientUserCHE, targetUserCHE, memberUsersCHE, memberUsernames

		RETURN { client_user_che: clientUserCHE, target_user_che: targetUserCHE, member_users_che: memberUsersCHE, member_usernames: memberUsernames } AS new_group_activity
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"target_user":     targetUser,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return RemoveUserActivity{}, fiber.ErrInternalServerError
	}

	newGact := modelHelpers.RKeyGet[RemoveUserActivity](res.Records, "new_group_activity")

	return newGact, nil
}

type UserJoinedActivity struct {
	GroupInfo       map[string]any `json:"-" db:"group_info"`
	ClientUserCHE   map[string]any `json:"-" db:"client_user_che"`
	MemberUsersCHE  map[string]any `json:"-" db:"member_users_che"`
	MemberUsernames []any          `json:"-" db:"member_usernames"`
}

func Join(ctx context.Context, groupId, clientUsername string) (UserJoinedActivity, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser:User{ username: $client_username }), (group:Group{ id: $group_id })
		WHERE NOT EXISTS { (clientUser)-[:IS_MEMBER_OF]->(group) }
			AND NOT EXISTS { (group)-[:REMOVED_USER]->(clientUser) }

		CREATE (cligact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You joined" })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, clientUser, group AS canNullG, cligact { .che_id, .che_type, .info } AS clientUserCHE
		OPTIONAL MATCH (canNullG)<-[:IS_MEMBER_OF]-(memberUser)
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(canNullG)

		LET memberUsernames = collect(memberUser.username),
			memberUsersCHE = reduce(m = {}, mu IN memberUsernames | m + [mu]: { che_id: randomUUID(), che_type: "group activity", info: $client_username + " joined" })

		FOREACH (mc IN collect(memberChat) | CREATE (:GroupChatEntry{ che_id: memberUsersCHE[mc.owner_username].che_id, che_type: memberUsersCHE[mc.owner_username].che_type, info: memberUsersCHE[mc.owner_username].info })-[:IN_GROUP_CHAT]->(mc))

		WITH group, clientUser, group AS canNullG, clientUser AS canNullCU, 
			clientUserCHE, memberUsersCHE, memberUsernames
		
		OPTIONAL MATCH (canNullCU)-[lgr:LEFT_GROUP]->(canNullG)

		DELETE lgr

		WITH group, clientUser, clientUserCHE, memberUsersCHE, memberUsernames
		CREATE (clientUser)-[:IS_MEMBER_OF { role: "member" }]->(group)
		MERGE (clientUser)-[:HAS_CHAT]->(clientChat:GroupChat{ owner_username: clientUser.username, group_id: $group_id })-[:WITH_GROUP]->(group)

		WITH DISTINCT group { .id, .name, .description, .picture_url, .created_at } AS groupInfo,
			clientUserCHE, memberUsersCHE, memberUsernames

		RETURN { group_info: groupInfo, client_user_che: clientUserCHE, member_users_che: memberUsersCHE, member_usernames: memberUsernames } AS new_group_activity
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return UserJoinedActivity{}, fiber.ErrInternalServerError
	}

	newGact := modelHelpers.RKeyGet[UserJoinedActivity](res.Records, "new_group_activity")

	return newGact, nil
}

type UserLeftActivity struct {
	ClientUserCHE   map[string]any `json:"-" db:"client_user_che"`
	MemberUsersCHE  map[string]any `json:"-" db:"member_users_che"`
	MemberUsernames []any          `json:"-" db:"member_usernames"`
}

func Leave(ctx context.Context, groupId, clientUsername string) (UserLeftActivity, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group:Group{ id: $group_id })<-[mem:IS_MEMBER_OF]-(clientUser:User{ username: $client_username }),
			(clientUser)-[:HAS_CHAT]->(clientChat)-[:WITH_GROUP]->(group)

		DELETE mem

		WITH group, clientUser, clientChat
		CREATE (clientUser)-[:LEFT_GROUP]->(group),
			(cligact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You left" })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, cligact { .che_id, .che_type, .info } AS clientUserCHE

		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser)
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(group)

		LET memberUsernames = collect(memberUser.username),
			memberUsersCHE = reduce(m = {}, mu IN memberUsernames | m + [mu]: { che_id: randomUUID(), che_type: "group activity", info: $client_username + " left" })

		FOREACH (mc IN collect(memberChat) | CREATE (:GroupChatEntry{ che_id: memberUsersCHE[mc.owner_username].che_id, che_type: memberUsersCHE[mc.owner_username].che_type, info: memberUsersCHE[mc.owner_username].info })-[:IN_GROUP_CHAT]->(mc))

		WITH DISTINCT clientUserCHE, memberUsersCHE, memberUsernames

		RETURN { client_user_che: clientUserCHE, member_users_che: memberUsersCHE, member_usernames: memberUsernames } AS new_group_activity
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return UserLeftActivity{}, fiber.ErrInternalServerError
	}

	newGact := modelHelpers.RKeyGet[UserLeftActivity](res.Records, "new_group_activity")

	return newGact, nil
}

type MakeUserAdminActivity struct {
	ClientUserCHE   map[string]any `json:"-" db:"client_user_che"`
	TargetUserCHE   map[string]any `json:"-" db:"target_user_che"`
	MemberUsersCHE  map[string]any `json:"-" db:"member_users_che"`
	MemberUsernames []any          `json:"-" db:"member_usernames"`
}

func MakeUserAdmin(ctx context.Context, groupId, clientUsername, targetUser string) (MakeUserAdminActivity, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
			(group)<-[mem:IS_MEMBER_OF { role: "member" }]-(targetUser:User{ username: $target_user })

		SET mem.role = "admin"

		CREATE (cligact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You made " + $target_user + " group admin" })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, targetUser, cligact { .che_id, .che_type, .info } AS clientUserCHE
		MATCH (targetUser)-[:HAS_CHAT]->(targetUserChat)-[:WITH_GROUP]->(group)

		CREATE (tugact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: $client_username + " made you group admin" })-[:IN_GROUP_CHAT]->(targetUserChat)

		WITH group, clientUserCHE, tugact { .che_id, .che_type, .info } AS targetUserCHE
		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE NOT memberUser.username IN [$client_username, $target_user])
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(group)

		LET memberUsernames = collect(memberUser.username),
			memberUsersCHE = reduce(m = {}, mu IN memberUsernames | m + [mu]: { che_id: randomUUID(), che_type: "group activity", info: $client_username + " made " + $target_user + " group admin" })

		FOREACH (mc IN collect(memberChat) | CREATE (:GroupChatEntry{ che_id: memberUsersCHE[mc.owner_username].che_id, che_type: memberUsersCHE[mc.owner_username].che_type, info: memberUsersCHE[mc.owner_username].info })-[:IN_GROUP_CHAT]->(mc))

		WITH DISTINCT clientUserCHE, targetUserCHE, memberUsersCHE, memberUsernames

		RETURN { client_user_che: clientUserCHE, target_user_che: targetUserCHE, member_users_che: memberUsersCHE, member_usernames: memberUsernames } AS new_group_activity
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"target_user":     targetUser,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return MakeUserAdminActivity{}, fiber.ErrInternalServerError
	}

	newGact := modelHelpers.RKeyGet[MakeUserAdminActivity](res.Records, "new_group_activity")

	return newGact, nil
}

type RemoveUserFromAdminsActivity struct {
	ClientUserCHE   map[string]any `json:"-" db:"client_user_che"`
	TargetUserCHE   map[string]any `json:"-" db:"target_user_che"`
	MemberUsersCHE  map[string]any `json:"-" db:"member_users_che"`
	MemberUsernames []any          `json:"-" db:"member_usernames"`
}

func RemoveUserFromAdmins(ctx context.Context, groupId, clientUsername, targetUser string) (RemoveUserFromAdminsActivity, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
			MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
				(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
				(group)<-[mem:IS_MEMBER_OF { role: "admin" }]-(targetUser:User{ username: $target_user })

			SET mem.role = "member"
				
			CREATE (cligact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: "You removed " + $target_user + " from group admins" })-[:IN_GROUP_CHAT]->(clientChat)

			WITH group, targetUser, cligact { .che_id, .che_type, .info } AS clientUserCHE
			MATCH (targetUser)-[:HAS_CHAT]->(targetUserChat)-[:WITH_GROUP]->(group)
			
			CREATE (tugact:GroupChatEntry{ che_id: randomUUID(), che_type: "group activity", info: $client_username + " removed you from group admins" })-[:IN_GROUP_CHAT]->(targetUserChat)

			WITH group, clientUserCHE, tugact { .che_id, .che_type, .info } AS targetUserCHE
			OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE NOT memberUser.username IN [$client_username, $target_user])
			OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(group)

			LET memberUsernames = collect(memberUser.username),
				memberUsersCHE = reduce(m = {}, mu IN memberUsernames | m + [mu]: { che_id: randomUUID(), che_type: "group activity", info: $client_username + " removed " + $target_user + " from group admins" })

			FOREACH (mc IN collect(memberChat) | CREATE (:GroupChatEntry{ che_id: memberUsersCHE[mc.owner_username].che_id, che_type: memberUsersCHE[mc.owner_username].che_type, info: memberUsersCHE[mc.owner_username].info })-[:IN_GROUP_CHAT]->(mc))

			WITH DISTINCT clientUserCHE, targetUserCHE, memberUsersCHE, memberUsernames

			RETURN { client_user_che: clientUserCHE, target_user_che: targetUserCHE, member_users_che: memberUsersCHE, member_usernames: memberUsernames } AS new_group_activity
			`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"target_user":     targetUser,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return RemoveUserFromAdminsActivity{}, fiber.ErrInternalServerError
	}

	newGact := modelHelpers.RKeyGet[RemoveUserFromAdminsActivity](res.Records, "new_group_activity")

	return newGact, nil
}

type NewMessage struct {
	Id                   string         `json:"id" db:"id"`
	ChatHistoryEntryType string         `json:"che_type" db:"che_type"`
	Content              map[string]any `json:"content" db:"content"`
	DeliveryStatus       string         `json:"delivery_status" db:"delivery_status"`
	CreatedAt            int64          `json:"created_at" db:"created_at"`
	Sender               any            `json:"sender" db:"sender"`
	ReplyTargetMsg       map[string]any `json:"reply_target_msg,omitempty" db:"reply_target_msg"`
	MemberUsernames      []any          `json:"-" db:"member_usernames"`
}

func SendMessage(ctx context.Context, clientUsername, groupId, msgContent string, at int64) (NewMessage, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser)
		WHERE EXISTS { (clientUser)-[:IS_MEMBER_OF]->(group) }

		CREATE (message:GroupMessage:GroupChatEntry{ id: randomUUID(), che_type: "message", content: $message_content, delivery_status: "sent", created_at: $at }),
			(clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_GROUP_CHAT]->(clientChat)
		
		WITH group, message

		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(group)

		FOREACH (memb IN collect({chat: memberChat, user: memberUser}) | CREATE (memb.user)-[:RECEIVES_MESSAGE]->(message)-[:IN_GROUP_CHAT]->(memb.chat))

		WITH DISTINCT message, collect(memberUser.username) AS member_usernames
		RETURN message { .*, content: apoc.convert.fromJsonMap(message.content), sender: $client_username, member_usernames } AS new_message
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"message_content": msgContent,
			"at":              at,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessage{}, fiber.ErrInternalServerError
	}

	newMessage := modelHelpers.RKeyGet[NewMessage](res.Records, "new_message")

	return newMessage, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, groupId, msgId string, deliveredAt int64) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientChat)<-[:IN_GROUP_CHAT]-(message:GroupMessage{ id: $message_id, delivery_status: "sent" })<-[:RECEIVES_MESSAGE]-(clientUser),

		CREATE (message)-[:DELIVERED_TO { at: $delivered_at }]->(clientUser)

		RETURN true AS done
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"message_id":      msgId,
			"delivered_at":    deliveredAt,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	return true, nil
}

func AckMessageRead(ctx context.Context, clientUsername, groupId, msgId string, readAt int64) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientChat)<-[:IN_GROUP_CHAT]-(message:GroupMessage{ id: $message_id } WHERE message.delivery_status <> "read")<-[:RECEIVES_MESSAGE]-(clientUser),

		CREATE (message)-[:READ_BY { at: $read_at }]->(clientUser)

		RETURN true AS done
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"message_id":      msgId,
			"read_at":         readAt,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	return true, nil
}

func ReplyToMessage(ctx context.Context, clientUsername, groupId, targetMsgId, msgContent string, at int64) (NewMessage, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser)
		WHERE EXISTS { (clientUser)-[:IS_MEMBER_OF]->(group) }

		MATCH (targetMsg:GroupMessage { id: $target_msg_id })

		CREATE (replyMsg:GroupMessage:GroupChatEntry{ id: randomUUID(), che_type: "message", content: $message_content, delivery_status: "sent", created_at: $at }),
			(clientUser)-[:SENDS_MESSAGE]->(replyMsg)-[:IN_GROUP_CHAT]->(clientChat),
			(replyMsg)-[:REPLIES_TO]->(targetMsg)
		
		WITH group, replyMsg, targetMsg

		MATCH (targetMsg)<-[:SENDS_MESSAGE]-(targetMsgSender)

		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(group)

		FOREACH (memb IN collect({chat: memberChat, user: memberUser}) | CREATE (memb.user)-[:RECEIVES_MESSAGE]->(replyMsg)-[:IN_GROUP_CHAT]->(memb.chat))

		WITH DISTINCT replyMsg, collect(memberUser.username) AS member_usernames,
			targetMsg { .id, content: apoc.convert.fromJsonMap(targetMsg.content), sender_user: targetMsgSender.username } AS reply_target_msg

		RETURN replyMsg { .*, content: apoc.convert.fromJsonMap(replyMsg.content), sender: $client_username, reply_target_msg, member_usernames } AS new_message
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"message_content": msgContent,
			"target_msg_id":   targetMsgId,
			"at":              at,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessage{}, fiber.ErrInternalServerError
	}

	newMessage := modelHelpers.RKeyGet[NewMessage](res.Records, "new_message")

	return newMessage, nil
}

type RxnToMessage struct {
	CHEId                string `json:"-" db:"che_id"`
	ChatHistoryEntryType string `json:"che_type" db:"che_type"`
	Emoji                string `json:"emoji" db:"emoji"`
	Reactor              any    `json:"reactor" db:"reactor"`
	MemberUsernames      []any  `json:"-" db:"member_usernames"`
}

func ReactToMessage(ctx context.Context, clientUsername, groupId, msgId, emoji string, at int64) (RxnToMessage, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser)
		WHERE EXISTS { (clientUser)-[:IS_MEMBER_OF]->(group) }

		MATCH (clientChat)<-[IN_GROUP_CHAT]-(message:GroupMessage{ id: $message_id })

		WITH group, clientUser, message, clientChat
		MERGE (msgrxn:GroupMessageReaction:GroupChatEntry{ reactor_username: clientUser.username, message_id: $message_id })
		ON CREATE
			SET msgrxn.che_id = randomUUID(),
				msgrxn.che_type = "reaction"

		SET msgrxn.emoji = $emoji, msgrxn.at = $at

		MERGE (clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)
		SET crxn.emoji = $emoji, crxn.at = $at

		MERGE (msgrxn)-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, msgrxn

		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)
		OPTIONAL MATCH (memberUser)-[:HAS_CHAT]->(memberChat)-[:WITH_GROUP]->(group)

		FOREACH (mc IN collect(memberChat) | CREATE (msgrxn)-[:IN_GROUP_CHAT]->(mc))

		LET member_usernames = collect(memberUser.username)

		RETURN msgrxn { .che_id, .che_type, .emoji, to_msg_id: $message_id, reactor: $client_username, member_usernames } rxn_to_msg
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"message_id":      msgId,
			"emoji":           emoji,
			"at":              at,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return RxnToMessage{}, fiber.ErrInternalServerError
	}

	rxnToMessage := modelHelpers.RKeyGet[RxnToMessage](res.Records, "rxn_to_msg")

	return rxnToMessage, nil
}

type RemovedRxnToMessage struct {
	CHEId           string `json:"-" db:"msgrxn_che_id"`
	MemberUsernames []any  `json:"-" db:"member_usernames"`
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, groupId, msgId string) (RemovedRxnToMessage, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser)
		WHERE EXISTS { (clientUser)-[:IS_MEMBER_OF]->(group) }

		MATCH (clientChat)<-[IN_GROUP_CHAT]-(message:GroupMessage{ id: $message_id }),
			(msgrxn:GroupMessageReaction:GroupChatEntry{ reactor_username: clientUser.username, message_id: message.id }),
			(clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)
		
		LET msgrxn_che_id = msgrxn.che_id

		DETACH DELETE msgrxn, crxn
		
		WITH group, msgrxn_che_id

		OPTIONAL MATCH (group)<-[:IS_MEMBER_OF]-(memberUser WHERE memberUser.username <> $client_username)

		LET member_usernames = collect(memberUser.username)

		RETURN DISTINCT { msgrxn_che_id, member_usernames } AS rmvd_rxn_to_msg
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"message_id":      msgId,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return RemovedRxnToMessage{}, fiber.ErrInternalServerError
	}

	rmvdRxnToMessage := modelHelpers.RKeyGet[RemovedRxnToMessage](res.Records, "rmvd_rxn_to_msg")

	return rmvdRxnToMessage, nil
}

func ChatHistory(ctx context.Context, clientUsername, groupId string, limit int, cursor float64) ([]UITypes.ChatHistoryEntry, error) {
	return nil, nil
}

func GroupInfo(ctx context.Context, groupId string) (UITypes.GroupInfo, error) {
	return UITypes.GroupInfo{}, nil
}

func GroupMembers(ctx context.Context, clientUsername, groupId string) ([]UITypes.GroupMemberSnippet, error) {
	return nil, nil
}
