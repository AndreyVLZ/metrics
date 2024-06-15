package model

const (
	TypeCountConst Type = iota
	TypeGaugeConst
)

type Type int8

const totalTypes = 2 // Кол-во типов для метрик.

// Получение типа из строки. Ошибка если тип не поддерживается.
func ParseType(typeStr string) (Type, error) {
	switch typeStr {
	case TypeCountConst.String():
		return TypeCountConst, nil
	case TypeGaugeConst.String():
		return TypeGaugeConst, nil
	default:
		return 0, errTypeNotSupport
	}
}

// Возващает имя для типа метрики.
func (t Type) String() string { return supportTypeMetric()[t] }

func supportTypeMetric() [totalTypes]string {
	return [totalTypes]string{
		"counter",
		"gauge",
	}
}
