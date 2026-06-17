package ctrl

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	ge "github.com/extrame/goblet/v2/error"
)

type Group struct {
	*Basic
	IgnoreCase bool
}

func (c *Group) MatchSuffix(suffix string) bool {
	return true
}

func (g *Group) String() string {
	var with = "with"
	if !g.IgnoreCase {
		with = "without"
	}
	return fmt.Sprintf("Group(%s) %s ignore case", g.Name, with)
}

func (g *Group) Parse(ctx Context) error {
	var method reflect.Value
	var name string

	const GetOptionButJustHasNormalMethod, GetOptionAndHasOption = 1, 2

	isOptions := 0
	var allowdMethods = []string{
		"Post",
		"Get",
	}

	var suffix, suffixWithSlash = ctx.Suffix(true)
	if len(suffix) > 0 && suffixWithSlash {

		matched, suffix, params := g.methods.Match(suffix, len(suffix))

		if matched != nil {
			if mc, ok := matched.Opt.(*MethodCaller); ok {
				method = mc.fn
				if method.IsValid() {
					ctx.SetSuffix(suffix)
					goto next
				}
			} else {
				slog.Info("matched not method caller", "matched", matched)
			}
		}

		if params != nil {
			ctx.SetPathParams(params)
		}
	}
	if !method.IsValid() {
		name = ctx.ReqMethod()
		switch name {
		case "OPTIONS":
			method = g.block.MethodByName("Options")
			name = "options"
			if method.IsValid() {
				isOptions = GetOptionAndHasOption
				allowdMethods = append(allowdMethods, "Options")
			} else {
				isOptions = GetOptionButJustHasNormalMethod
				method = g.block.MethodByName("Post") //for test valid or not
			}
		default:
			methodName := name[:1] + strings.ToLower(name[1:])
			method = g.block.MethodByName(methodName)
		}
	}

next:
	if isOptions == GetOptionButJustHasNormalMethod {
		// transform to upper case
		var upperMethods = make([]string, len(allowdMethods))
		for i, v := range allowdMethods {
			upperMethods[i] = strings.ToUpper(v)
		}
		ctx.SetHeader("Allow", strings.Join(upperMethods, ", "))
		var found = false
		for _, v := range allowdMethods {
			method := g.block.MethodByName(v)
			if method.IsValid() {
				found = true
				break
			}
		}
		if !found {
			return ge.NOSUCHROUTER("")
		}
		if g.tryPre("options", ctx) {
			ctx.RespondOK()
		}
		return nil
	}
	if !method.IsValid() {
		return ge.NOSUCHROUTER("")
	} else {
		ctx.RenderAs(name)

		if g.tryPre(name, ctx) {
			results, typ := callMethod(method, ctx)
			return checkResult(results, typ, ctx)
		}

	}
	return nil

}
