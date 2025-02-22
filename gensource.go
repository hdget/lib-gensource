package gensource

import (
	"bytes"
	"github.com/dave/jennifer/jen"
	reflectUtils "github.com/hdget/hdutils/reflect"
	"os"
)

type genSourceUtilsImpl struct {
	f *jen.File
}

func New(pkgName string, imports map[string]string) Utils {
	f := jen.NewFile(pkgName)
	f.ImportNames(imports)
	return &genSourceUtilsImpl{
		f: f,
	}
}

// AddMethod 增加方法
func (h *genSourceUtilsImpl) AddMethod(receiver, methodName string, params, results []string, body []jen.Code) Utils {
	statement := jen.Func().Params(jen.Op("*").Id(receiver)).Id(methodName).Params()
	for _, result := range results {
		switch result {
		case "any":
			statement = statement.Any()
		case "string":
			statement = statement.String()
		case "panic":
			statement = statement.Error()

		}
	}

	h.f.Add([]jen.Code{statement}...).Block(
		body...,
	)
	return h
}

func (h *genSourceUtilsImpl) Save(destFile string) error {
	// 保存数据
	buf := &bytes.Buffer{}
	err := h.f.Render(buf)
	if err != nil {
		return err
	}

	err = os.WriteFile(destFile, buf.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// DeclareSliceVar 声明Slice变量并赋值
func (h *genSourceUtilsImpl) DeclareSliceVar(varName string, valueImportPath string, values []any) Utils {
	valueCodes := h.getSliceValuesCodes(valueImportPath, values)
	h.f.Var().Id(varName).Op("=").Add(valueCodes...)
	return h
}

// newSliceVar 获取Slice值的代码
func (h *genSourceUtilsImpl) getSliceValuesCodes(valueImportPath string, values []any) []jen.Code {
	if len(values) == 0 {
		return nil
	}

	var isValuePointer bool
	var valueTypeName string
	valueCodes := make([]jen.Code, 0)
	for _, v := range values {
		// 检视每个值的信息，包括名字，值，类型
		valueInfo := reflectUtils.InspectValue(v)

		// values中的值的类型都是一样，获取值的类型名字
		if valueTypeName == "" {
			valueTypeName = valueInfo.Name
		}

		// 判断值是否是指针类型
		isValuePointer = valueInfo.IsPointer

		// 值有可能是struct, 遍历值的所有struct field, 设置为值
		valueCodes = append(valueCodes, h.getValuesCodes(v)...)
	}

	var statement *jen.Statement
	if isValuePointer {
		statement = jen.Index().Op("*").Qual(valueImportPath, valueTypeName).Values(valueCodes...)
	} else {
		statement = jen.Index().Qual(valueImportPath, valueTypeName).Values(valueCodes...)
	}
	return []jen.Code{statement}
}

// getValuesCodes 获取value对应的jen.Codes值
func (h *genSourceUtilsImpl) getValuesCodes(v any) []jen.Code {
	val := reflectUtils.InspectValue(v)
	var codes []jen.Code
	switch val.Kind {
	case "struct":
		return h.getStructValueCodes(val, v)
	case "slice":
		return h.getSliceValueCodes(val, v)
	default:
		codes = []jen.Code{jen.Lit(v)}
	}
	return codes
}

// getStructValueCodes 获取struct value的jen codes
func (h *genSourceUtilsImpl) getStructValueCodes(val *reflectUtils.Value, v any) []jen.Code {
	codes := make([]jen.Code, 0)
	// 值有可能是struct, 遍历值的所有struct field, 设置为值
	codes = append(codes, jen.Values(jen.DictFunc(func(d jen.Dict) {
		for _, item := range val.Items {
			itemCodes := h.getValuesCodes(item.Value)
			if len(itemCodes) == 0 {
				d[jen.Id(item.Name)] = jen.Nil()
			} else {
				d[jen.Id(item.Name)] = itemCodes[0]
			}
		}
	})))

	return codes
}

// getSliceValueCodes 获取slice value的jen codes
func (h *genSourceUtilsImpl) getSliceValueCodes(val *reflectUtils.Value, v any) []jen.Code {
	// 获取slice item的类型
	var itemKind string
	itemCodes := make([]jen.Code, 0)
	for _, item := range val.Items {
		if itemKind == "" {
			itemKind = item.Kind
		}
		itemCodes = append(itemCodes, h.getValuesCodes(item.Value)...)
	}

	if len(itemCodes) == 0 {
		return []jen.Code{jen.Nil()}
	}

	codes := make([]jen.Code, 0)
	switch itemKind {
	case "text":
		codes = append(codes, jen.Index().String().Values(itemCodes...))
	case "int":
		codes = append(codes, jen.Index().Int().Values(itemCodes...))
	case "int64":
		codes = append(codes, jen.Index().Int64().Values(itemCodes...))
	case "int32":
		codes = append(codes, jen.Index().Int32().Values(itemCodes...))
	case "float64":
		codes = append(codes, jen.Index().Float64().Values(itemCodes...))
	case "float32":
		codes = append(codes, jen.Index().Float32().Values(itemCodes...))
	default:
		codes = append(codes, jen.Index().Values(itemCodes...))
	}

	return codes
}
