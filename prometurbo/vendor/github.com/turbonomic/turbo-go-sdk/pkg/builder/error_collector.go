package builder

import (
	"fmt"
	"strings"
)

type ErrorCollector []error

func (ec *ErrorCollector) Count() int {
	if ec == nil {
		return 0
	}
	return len(*ec)
}

func (ec *ErrorCollector) Collect(err error) {
	if err != nil {
		*ec = append(*ec, err)
	}
}

func (ec *ErrorCollector) CollectAll(errList []error) {
	for i, _ := range errList {
		err := errList[i]
		if err != nil {
			*ec = append(*ec, err)
		}
	}
}

func (ec *ErrorCollector) Error() string {
	var errorStr []string
	errorStr = append(errorStr, "GroupBuilder errors:")
	for i, err := range *ec {
		errorStr = append(errorStr, fmt.Sprintf("Error %d: %s", i, err.Error()))
	}
	return strings.Join(errorStr, " ")
}
