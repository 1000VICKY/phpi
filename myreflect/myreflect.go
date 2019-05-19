package myreflect

import "errors"
import "reflect"

/**
 * PHPのget_class_methodsを再現
 * 指定したオブジェクトが持つpublicなメソッド一覧を取得する
 *
 */
func GetObjectMethods(s interface{}) ([]string, error) {
	var typeObj reflect.Type = reflect.TypeOf(s)

	var valueObj reflect.Value = reflect.New(typeObj).Elem()
	// 指定した型が持つpublicなメソッドの件数を取得
	var methodCount int = valueObj.NumMethod()
	var methodList []string = make([]string, methodCount)

	var typeInLoop reflect.Type
	var methodInLoop reflect.Method
	for i := 0; i < methodCount; i++ {
		// reflect.Type型
		typeInLoop = valueObj.Type()
		// reflect.Method型
		methodInLoop = typeInLoop.Method(i)
		methodList = append(methodList, methodInLoop.Name)
	}
	if len(methodList) > 0 {
		return methodList, nil
	} else {
		return make([]string, 0), errors.New("対象のオブジェクトはメソッドを持っていません。")
	}
}
