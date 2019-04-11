package myreflect;

import "errors";
import "reflect";
/**
 * 指定したオブジェクトが持つpublicなメソッド一覧を取得する
 *
 */
func GetObjectMethods(s interface{}) ([]string, error) {
    var t reflect.Type = reflect.TypeOf(s);
    var v reflect.Value = reflect.New(t).Elem();
    var methodCount int = v.NumMethod();
    var methodList []string = make([]string, methodCount);
    for i := 0; i < methodCount; i++ {
        typeInLoop := v.Type();
        methodInLoop := typeInLoop.Method(i);
        methodList = append(methodList, methodInLoop.Name);
    }
    if (len(methodList) > 0) {
        return methodList, nil;
    } else {
        return make([]string, 0), errors.New("対象のオブジェクトはメソッドを持っていません。");
    }
}


