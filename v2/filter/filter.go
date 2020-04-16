package filter

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/pixeltopic/sayori/v2/context"
)

// FmtErrDefault is the default format function to convert failing Filter(s) into an error string
func FmtErrDefault(f Filter) string {
	return fmt.Sprintf("filter fail code '%d'", f)
}

// Filter represents a condition that prevents a `Command` or `Event` from firing.
// Only use the given const filters which range from 2^0 to 2^4.
type Filter int

// AsError returns the filter as an error
func (f Filter) AsError() error {
	return &Error{
		f:      f,
		reason: FmtErrDefault(f),
	}
}

// Error is an error that has a failing Filter attached
type Error struct {
	f      Filter
	reason string
}

// Filter returns the violated filter
func (e *Error) Filter() Filter {
	return e.f
}

func (e *Error) Error() string {
	return e.reason
}

const (
	// MsgFromSelf ignores handling of messages from the bot itself
	MsgFromSelf Filter = 1 << iota
	// MsgFromBot ignores handling of messages from all bots. Encapsulates MsgFromSelf.
	MsgFromBot
	// MsgFromWebhook ignores handling of messages from webhooks
	MsgFromWebhook
	// MsgNoContent ignores handling if messages without a text body. (e.g. messages with only attachments)
	MsgNoContent
	// MsgIsPrivate ignores handling if the message is sent in a DM
	MsgIsPrivate
	// MsgIsGuildText ignores handling if the message is sent from a guild text channel
	MsgIsGuildText
)

// New generates a Filter bitset given filters and performing a bitwise "or" on all of them
func New(filters ...Filter) Filter {
	var filter Filter
	for _, f := range filters {
		filter = filter | f
	}
	return filter
}

// Contains returns true if the given filter is part of the current filter
func (f Filter) Contains(filter Filter) bool {
	return f.ignores(filter)
}

// ignores returns true if the given filter is applied to the current filter
func (f Filter) ignores(filter Filter) bool {
	return f&filter == filter
}

// Validate inspects context and determines if it should be processed or not.
//
// returns true if allowed with a zero value Filter, or false with all failing Filters combined with a bitwise `or`.
//
// if ctx.Msg or ctx.Ses is nil, will return false with a zero value Filter.
func (f Filter) Validate(ctx *context.Context) (bool, Filter) {
	var failed Filter
	if ctx.Msg == nil || ctx.Ses == nil {
		return false, failed
	}

	var (
		contentLen = len(ctx.Msg.Content)
	)

	if f.ignores(MsgFromSelf) {
		switch {
		case ctx.Msg.Author == nil:
			fallthrough
		case ctx.Ses.State == nil:
			fallthrough
		case ctx.Ses.State.User == nil:
			return false, Filter(0)
		case ctx.Msg.Author.ID == ctx.Ses.State.User.ID:
			failed = failed | MsgFromSelf
		}
	}

	if f.ignores(MsgFromBot) {
		switch {
		case ctx.Msg.Author == nil:
			return false, Filter(0)
		case ctx.Msg.Author.Bot:
			failed = failed | MsgFromBot
		}
	}

	if f.ignores(MsgFromWebhook) && len(ctx.Msg.WebhookID) != 0 {
		failed = failed | MsgFromWebhook
	}

	if f.ignores(MsgNoContent) && contentLen == 0 {
		failed = failed | MsgNoContent
	}
	if f.ignores(MsgIsPrivate) && comesFromDM(ctx.Ses, ctx.Msg) {
		failed = failed | MsgIsPrivate
	}
	if f.ignores(MsgIsGuildText) && comesFromGuild(ctx.Ses, ctx.Msg) {
		failed = failed | MsgIsGuildText
	}

	return failed == 0, failed
}

// comesFromDM returns true if a message comes from a DM channel
func comesFromDM(s *discordgo.Session, m *discordgo.Message) bool {

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return len(m.GuildID) == 0 // final fallback to check if DM
		}
	}

	return channel.Type == discordgo.ChannelTypeDM
}

// comesFromGuild determines if a message was sent in a text channel within a guild
func comesFromGuild(s *discordgo.Session, m *discordgo.Message) bool {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return len(m.GuildID) != 0 // final fallback to check if a text channel in guild
		}
	}

	return channel.Type == discordgo.ChannelTypeGuildText
}
