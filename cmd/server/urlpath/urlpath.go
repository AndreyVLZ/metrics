package urlpath

import "errors"

const (
	setTypeConst = iota
	setNameConst
	setValueConst
)

var ErrNoCorrectURLPath error = errors.New("no correct url path")

type GetURLPath struct {
	typeStr string
	name    string
}

func NewGetURLPath(arr ...string) *GetURLPath {
	getURLPath := new(GetURLPath)
	for i := range arr {
		getURLPath.set(i, arr[i])
	}

	return getURLPath
}

func (urlPath *GetURLPath) Type() string { return urlPath.typeStr }
func (urlPath *GetURLPath) Name() string { return urlPath.name }

func (urlPath *GetURLPath) Validate() error {
	if urlPath.name == "" || urlPath.typeStr == "" {
		return ErrNoCorrectURLPath
	}

	return nil
}

func (urlPath *GetURLPath) set(key int, val string) {
	switch key {
	case setTypeConst:
		urlPath.typeStr = val
	case setNameConst:
		urlPath.name = val
	}
}

type UpdateURLPath struct {
	*GetURLPath
	value string
}

func NewUpdateURLPath(arr ...string) *UpdateURLPath {
	updateURLPath := &UpdateURLPath{GetURLPath: new(GetURLPath)}
	for i := range arr {
		updateURLPath.set(i, arr[i])
	}

	return updateURLPath
}

func (urlPath *UpdateURLPath) Value() string { return urlPath.value }

func (urlPath *UpdateURLPath) Validate() error {
	if urlPath.value == "" || urlPath.GetURLPath.Validate() != nil {
		return ErrNoCorrectURLPath
	}

	return nil
}

func (urlPath *UpdateURLPath) set(key int, val string) {
	switch key {
	case setTypeConst, setNameConst:
		urlPath.GetURLPath.set(key, val)
	case setValueConst:
		urlPath.value = val
	}
}
