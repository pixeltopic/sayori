package v2

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type ctxKey int

const (
	ctxSesKey ctxKey = iota
	ctxMsgKey
	ctxPrefKey
	ctxAliasKey
	ctxArgsKey
	ctxCmdErrKey
)

// CmdContext is an aux structure for storing invocation values extracted from a given Context to reduce boilerplate.
//
// This need not be manually initialized; simply call CmdFromContext.
type CmdContext struct {
	Ses    *discordgo.Session
	Msg    *discordgo.Message
	Prefix string
	Alias  []string
	Args   []string
	Err    error
}

// CmdFromContext derives all Command invocation values from given Context.
func CmdFromContext(ctx context.Context) *CmdContext {
	return &CmdContext{
		Ses:    GetSes(ctx),
		Msg:    GetMsg(ctx),
		Prefix: GetPrefix(ctx),
		Alias:  GetAlias(ctx),
		Args:   GetArgs(ctx),
		Err:    GetErr(ctx),
	}
}

// WithSes attaches a Discord Session to Context.
func WithSes(ctx context.Context, ses *discordgo.Session) context.Context {
	return context.WithValue(ctx, ctxSesKey, ses)
}

// GetSes returns a Discord Session from Context. If Session not present, returns nil.
func GetSes(ctx context.Context) *discordgo.Session {
	ses, ok := ctx.Value(ctxSesKey).(*discordgo.Session)
	if !ok {
		return nil
	}
	return ses
}

// WithMsg attaches a Discord Message to Context.
func WithMsg(ctx context.Context, msg *discordgo.Message) context.Context {
	return context.WithValue(ctx, ctxMsgKey, msg)
}

// GetMsg returns a Discord Message from Context. If Message not present, returns nil.
func GetMsg(ctx context.Context) *discordgo.Message {
	msg, ok := ctx.Value(ctxMsgKey).(*discordgo.Message)
	if !ok {
		return nil
	}
	return msg
}

// WithPrefix attaches a Command Prefix to Context.
func WithPrefix(ctx context.Context, prefix string) context.Context {
	return context.WithValue(ctx, ctxPrefKey, prefix)
}

// GetPrefix returns a Command Prefix from Context.
func GetPrefix(ctx context.Context) string {
	v, ok := ctx.Value(ctxPrefKey).(string)
	if !ok {
		return ""
	}
	return v
}

// WithAlias attaches Command Aliases to Context.
func WithAlias(ctx context.Context, aliases []string) context.Context {
	return context.WithValue(ctx, ctxAliasKey, aliases)
}

// GetAlias returns Command Aliases from Context.
func GetAlias(ctx context.Context) []string {
	v, ok := ctx.Value(ctxAliasKey).([]string)
	if !ok {
		return []string{}
	}
	return v
}

// WithArgs attaches Command Args to Context.
func WithArgs(ctx context.Context, args []string) context.Context {
	return context.WithValue(ctx, ctxArgsKey, args)
}

// GetArgs returns Command Args from Context.
func GetArgs(ctx context.Context) []string {
	v, ok := ctx.Value(ctxArgsKey).([]string)
	if !ok {
		return []string{}
	}
	return v
}

// WithErr attaches a Command Err to Context.
func WithErr(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ctxCmdErrKey, err)
}

// GetErr returns a Command Err from Context, not to be confused with the default Context error.
func GetErr(ctx context.Context) error {
	v, ok := ctx.Value(ctxCmdErrKey).(error)
	if !ok {
		return nil
	}
	return v
}
