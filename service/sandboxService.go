package service

import (
	"fmt"
	"gin-app/db"
	"gin-app/param"
)

func GetSandboxList(req param.ReqSandbox) ([]param.ResSandbox, error) {

	list, err := db.GetSandboxList(map[string]interface{}{"KeyCl": req.KeyCl})
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		return nil, err
	}

	resList := make([]param.ResSandbox, 0, len(list))
	for _, v := range list {
		var jsonVal string
		if v.JsonCl.Valid {
			jsonVal = v.JsonCl.String()
		}

		e := param.ResSandbox{
			KeyCl:       v.KeyCl,
			IntCl:       v.IntCl,
			StrCl:       v.StrCl,
			BoolCl:      v.BoolCl,
			ByteCl:      v.ByteCl,
			DateCl:      v.DateCl.Date,
			TimeStampCl: v.TimeStampCl.Time,
			JsonCl:      jsonVal,
		}
		resList = append(resList, e)

	}

	return resList, nil

}
