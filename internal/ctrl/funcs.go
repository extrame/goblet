package ctrl

import (
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	ge "github.com/extrame/goblet/error"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func callMethod(method reflect.Value, ctx Context) ([]reflect.Value, reflect.Type) {
	typ := method.Type()
	rvArgs := make([]reflect.Value, typ.NumIn())
	var i = 0
	var suffix = ctx.Suffix()

	if len(suffix) > 0 && suffix[0] == '/' {
		suffix = suffix[1:]
	}

	for ; i < typ.NumIn(); i++ {
		argT := typ.In(i)
		var kind = argT.Kind()
		if kind == reflect.String || (kind >= reflect.Int && kind <= reflect.Int64) {
			args := strings.SplitN(suffix, "/", 2)
			var newV = reflect.New(argT)

			if kind == reflect.String {
				newV.Elem().SetString(args[0])
			} else {
				iValue, _ := strconv.ParseInt(args[0], 10, 64)
				newV.Elem().SetInt(iValue)
			}

			rvArgs[i] = newV.Elem()

			if len(args) >= 2 {
				suffix = args[1]
			} else {
				suffix = ""
			}
		} else if kind == reflect.Slice && argT.Elem().Kind() == reflect.String {
			args := strings.SplitN(suffix, "/", -1)
			rvArgs[i] = reflect.ValueOf(args)
			i++
			break
		} else {
			break
		}
	}

	rvArgs[i] = reflect.ValueOf(ctx)
	i++

	if i < typ.NumIn() {
		var typArg = typ.In(i)

		// 检查参数类型是否为map类型
		if typArg.Kind() == reflect.Map {
			// 对于map类型，直接从请求体中解析JSON
			// 创建对应类型的map实例
			mapType := reflect.MapOf(typArg.Key(), typArg.Elem())
			mapValue := reflect.New(mapType).Elem()

			// 创建指向map的指针用于Fill
			mapPtr := reflect.New(mapType).Interface()

			if err := ctx.Fill(mapPtr); err == nil {
				// 填充成功，将解析的map赋值给参数
				rvArgs[i] = reflect.ValueOf(mapPtr).Elem()
			} else {
				// 如果填充失败，返回空的map
				rvArgs[i] = mapValue
			}
		} else {
			// 原有的结构体参数处理逻辑
			var newV reflect.Value
			if typArg.Kind() == reflect.Ptr {
				slog.Info("parse arguments", "typArg", typArg, "pointer", typArg.Elem())
				newV = reflect.New(typ.In(i).Elem())
			} else {
				slog.Info("parse arguments", "typArg", typArg)
				newV = reflect.New(typ.In(i))
			}

			if err := ctx.Fill(newV.Interface()); err != nil {
				slog.Error("parse arguments error", "error", err)
			}
			if typArg.Kind() == reflect.Ptr {
				rvArgs[i] = newV
			} else {
				rvArgs[i] = newV.Elem()
			}
		}
	}

	return method.Call(rvArgs), typ
}

func checkResult(results []reflect.Value, typ reflect.Type, ctx Context) error {
	// status_code is not setted
	if len(results) > 0 && ctx.NotRendered() {
		for i := len(results); i > 0; i-- {
			var ires = results[i-1].Interface()
			var ti = typ.Out(i - 1)
			ok := ti.Implements(errorInterface)
			if !ok {
				ctx.Respond(ires)
				return nil
			} else if ok && !results[i-1].IsNil() {
				ierr := ires.(error)
				if ge.IsNoSuchRouter(ierr) {
					//程序中要求进行Nosuchrouter处理的，直接返回给路由
					return ierr
				}
				//优先按最后一个error参数返回错误，符合主流编程习惯
				ctx.RespondError(ires.(error))
				return nil
			}
		}
		ctx.RespondOK()
	}
	return nil
}
