# Progress
- Implement the sending of different message types
- 

---

### Message Content Structure
> This is just for documentation purposes. It's not implemented
```go
type MessageContent interface {
  Print()
}


type Message struct {
  SenderId int
  ChatId int
  Content MessageContent
}

type Text struct {
  Type string
  TextContent string
}
func(txt Text) Print() {}


type Voice struct {
  Type string
  Url string
  Duration time.Duration
}
func(voi Voice) Print() {}


type Image struct {
  Type string
  MimeType string
  Url string
  Caption string
  Size int
}
func(img Image) Print() {}

type Audio struct {
  Type string
  MimeType string
  Url string
  Caption string
  Size int
}
func(aud Audio) Print() {}

type Video struct {
  Type string
  MimeType string
  Url string
  Caption string
  Size int
}
func(vid Video) Print() {}

type FileAttachment struct {
  Type string
  MimeType string
  Url string
  Name string
  Size int
}
func(fat FileAttachment) Print() {}
```

### Recent Activity Structure
```go
type MessageActivity struct {
  SenderUsername string
  MessageType    string
  DeliveryStatus string
  // the text/caption of the message type if exists
  TextContent    string
}

type ReactionActivity struct {
  ReactorUsername string
  Reaction        rune
  MessageType     string
  // the text/caption of the message type if exists
  TextContent     string 
}

type GroupManagemenentActivity struct {
  // one of a defined set of activity types
  // {user_added, user_removed, user_left, user_joined, make_group_admin, removed_from_group_admins, user_changed_their_info, group_created}
  
  Type string 

  // other fields based on the activity type based on a defined structure
}

type GroupCreated struct {
  Creator         string
  GroupName       string
}

type UsersAdded struct {
  AddedBy           string
  NewUsers        []string
}

type UserRemoved struct {
  RemovedBy string
  Username string
}

type UserJoined struct {
  Username  string
}

type UserLeft struct {
  Username  string
}

type UserMadeGroupAdmin struct {
  MadeBy    string
  Username string
}

type UserRemovedFromGroupAdmins struct {
  RemovedBy    string
  Username string
}

type AdminChangedGroupName struct {
  AdminName string
  NewGroupName string
}

type AdminChangedGroupDescription struct {
  AdminName string
  NewGroupDescription string
}
```