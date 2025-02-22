package gensource

import "github.com/dave/jennifer/jen"

type Utils interface {
	DeclareSliceVar(varName string, valueImportPath string, values []any) Utils             // 声明变量并给变量赋值slice
	AddMethod(receiver, methodName string, params, results []string, body []jen.Code) Utils // 增加方法
	Save(destFile string) error                                                             // 保存文件
}
