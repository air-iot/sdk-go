/**
 * @Author: ZhangQiang
 * @Description:
 * @File:  type
 * @Version: 1.0.0
 * @Date: 2020/8/4 14:41
 */
package api

type AuthToken struct {
	Expires int64  `json:"expires"`
	Token   string `json:"token"`
}

// 采集数据
type (
	RealTimeData struct {
		TagId string      `json:"tagId"`
		Uid   string      `json:"uid"`
		Time  int64       `json:"time"`
		Value interface{} `json:"value"`
	}

	QueryData struct {
		Results []Results `json:"results"`
	}

	Series struct {
		Name    string          `json:"name"`
		Columns []string        `json:"columns"`
		Values  [][]interface{} `json:"values"`
	}

	Results struct {
		Series []Series `json:"series"`
	}
)
