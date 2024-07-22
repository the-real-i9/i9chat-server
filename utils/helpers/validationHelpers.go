package helpers

import (
	"i9chat/utils/appTypes"
	"regexp"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func MsgContentRule(msgType string) []validation.Rule {
	return []validation.Rule{
		validation.Required,
		validation.Map(
			validation.Key("type",
				validation.Required,
				validation.In("text", "voice", "audio", "video", "image", "file").Error("invalid message type"),
			),
			validation.Key("props",
				validation.Required,
				validation.Map(
					validation.Key("textContent", validation.When(msgType != "text", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
					validation.Key("data", validation.When(msgType == "text", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
					validation.Key("duration", validation.When(msgType != "voice", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
					validation.Key("mimeType",
						validation.When(slices.Contains([]string{"voice", "text"}, msgType), validation.Nil.Error("invalid property for the specified type")).Else(
							validation.Required,
							validation.Match(regexp.MustCompile("^[[:alnum:]!#$&^_.+-]+/[[:alnum:]!#$&^_.+-]+(?:;[[:blank:]]*[[:alnum:]!#$&^_.+-]+=[[:alnum:]!#$&^_.+-]+)*$")),
						),
					),
					validation.Key("size",
						validation.When(slices.Contains([]string{"voice", "text"}, msgType), validation.Nil.Error("invalid property for the specified type")).Else(
							validation.Required,
							validation.Min(1).Error("size cannot be zero bytes"),
							validation.Max(10*1024*1024).Error("maximum bytes of 10mb exceeded"),
						),
					),
					validation.Key("caption", validation.When(!slices.Contains([]string{"image", "auido", "video"}, msgType), validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
					validation.Key("ext", validation.When(msgType != "file", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
					validation.Key("name", validation.When(msgType != "file", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
				),
			),
		),
	}
}

var UserSliceRule validation.RuleFunc = func(value any) error {
	user := value.([]appTypes.String)

	if err := validation.Validate(user[0], validation.Required, is.Int.Error("invalid non-integer value")); err != nil {
		return err
	}

	if err := validation.Validate(user[1], validation.Required); err != nil {
		return err
	}

	return nil
}
