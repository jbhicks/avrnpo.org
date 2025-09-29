package actions

import (
	"net/http"
)

func (as *ActionSuite) Test_AdminSimple() {
	res := as.HTML("/admin/").Get()
	as.Equal(http.StatusFound, res.Code)
}
