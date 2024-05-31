package model

const (
	TypeCountConst Type = iota
	TypeGaugeConst
)

type Type int8

const totalTypes = 2

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

func (t Type) String() string { return supportTypeMetric()[t] }

func supportTypeMetric() [totalTypes]string {
	return [totalTypes]string{
		"counter",
		"gauge",
	}
}
