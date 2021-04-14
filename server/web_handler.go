package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"reflect"
	"time"
)

/*
"barcode" : "xxxxx",
		"goodsname" : "xxxxx",
		"quantity" 100,
		"cost" : 123.23,
		"sale_price" 123.55
*/

type GoodsInfo struct {
	BarCode    string  `json:"barcode"`
	GoodsName  string  `json:"goodsname"`
	Quantity   int     `json:"quantity"`
	Cost       float64 `json:"cost"`
	SalesPrice float64 `json:"sale_price"`
}

type QueryGoodsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []*GoodsInfo
}

type LocationInfo struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	UpdateTime string `json:"updatetime"`
}

type QueryLocationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []*LocationInfo
}

type WebHandlers struct {
	db *MysqlHandle
}

func NewWebHandlers(db *MysqlHandle) *WebHandlers {
	return &WebHandlers{
		db: db,
	}
}

/*
/api/inputgoods
request:
{
	"barcode" : "1234567890",
    "goodsname" : "测试商品名字1",
    "expiredate" : "2021-09-10",
	"quantity" :1000,
	"cost" : 100,
	"saleprice": 123.55,
	"location" : "门店1"
}
response :
{
	"code" : 0 // 0 ok, or error
	"message" : "xxxxx",
}
*/
func (h *WebHandlers) HandleInputGoods(resp http.ResponseWriter, req *http.Request) {
	LogInfo("handle input goods request")
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		SendResponse(resp, -1, "read request failed")
		return
	}
	reqData := make(map[string]interface{})
	LogInfo("REQDATA is %s", string(b))
	if err2 := json.Unmarshal(b, &reqData); err2 != nil {
		SendResponse(resp, -2, "parse request failed")
		LogError("parse request failed, %s  %s", err2.Error(), string(b))
		return
	}
	barCode, _ := reqData["barcode"].(string)
	if len(barCode) == 0 {
		SendResponse(resp, -3, "barcode is empty")
		return
	}
	query := fmt.Sprintf("select * from sales_storage.goods_info where bar_code = '%s'", barCode)
	rows := h.db.ExecQuery(query)
	if rows == nil {
		SendResponse(resp, -4, "query db failed")
		return
	}
	var bc, name, updateTime string
	var id int
	var price float64
	for rows.Next() {
		rows.Scan(&id, &name, &bc, &price, &updateTime)
		break
	}
	rows.Close()
	// goods not exists in goods info table, so it is a new goods category
	goodsName, _ := reqData["goodsname"].(string)
	salesPrice, _ := reqData["saleprice"].(float64)
	if len(bc) == 0 {
		insertSql := fmt.Sprintf("insert into sales_storage.goods_info(goods_name, bar_code, sales_price) values ('%s', '%s', %f)", goodsName, barCode, salesPrice)
		if err1 := h.db.Exec(insertSql); err1 != nil {
			LogError("insert to sales_storage.goods_info failed %s", err1.Error())
			SendResponse(resp, -5, "insert to db failed")
			return
		}
		bc = barCode
	} else {
		if name != goodsName || math.Abs(salesPrice-price) > 0.01 {
			updateSql := fmt.Sprintf("UPDATE sales_storage.goods_info set goods_name = '%s' , sales_price = %f, update_time = now()", goodsName, salesPrice)
			if err1 := h.db.Exec(updateSql); err1 != nil {
				LogError("update to sales_storage.goods_info failed %s", err1.Error())
				SendResponse(resp, -6, "update db failed")
				return
			}
		}
	}
	querySql := fmt.Sprintf("select id from sales_storage.goods_info where bar_code = '%s'", bc)
	rows2 := h.db.ExecQuery(querySql)
	if rows2 == nil {
		LogError("get goods info  failed")
		SendResponse(resp, -6, "get goods info failed")
		return
	}
	for rows2.Next() {
		rows2.Scan(&id)
		break
	}
	rows2.Close()
	location, _ := reqData["location"].(string)
	querySql = fmt.Sprintf("select id from sales_storage.goods_location where name = '%s'", location)
	rows3 := h.db.ExecQuery(querySql)
	if rows3 == nil {
		LogError("get goods location  failed")
		SendResponse(resp, -7, "get goods location failed")
		return
	}
	var locationID int64
	for rows3.Next() {
		rows3.Scan(&locationID)
		break
	}
	rows3.Close()
	if locationID == 0 {
		insertSql := fmt.Sprintf("insert into sales_storage.goods_location(name) values('%s')", location)
		err4, insertID := h.db.ExecWithInsertID(insertSql)
		if err4 != nil {
			LogError("insert location failed, %s", err4.Error())
			SendResponse(resp, -8, "save location failed")
			return
		}
		locationID = insertID
	}
	expireDate, _ := reqData["expiredate"].(string)
	LogInfo("%v", reflect.TypeOf(reqData["quantity"]))
	qty, _ := reqData["quantity"].(float64)

	cost, _ := reqData["cost"].(float64)
	insertSql := fmt.Sprintf("insert into sales_storage.goods_input(batch_no, storage_id, goods_id, expire_date, quantity, cost) values ('%s', %d, %d, '%s', %0.2f, %f)",
		h.getBatchNo(), 0, id, expireDate, qty, cost)
	LogInfo(insertSql)
	err5, inputID := h.db.ExecWithInsertID(insertSql)
	if err5 != nil {
		LogError("insert goods input failed, %s", err5.Error())
		SendResponse(resp, -10, "insert goods input failed")
		return
	}
	LogInfo("update storage begin")
	h.updateStorage(inputID, id, int(qty), cost)
	LogInfo("update storage end")
	SendResponse(resp, 0, "success")

}

func (h *WebHandlers) getBatchNo() string {
	return time.Now().Format("20060102150405")
}

func (h *WebHandlers) updateStorage(inputID int64, goodsID int, qty int, cost float64) {
	LogInfo("update storage with %d %d %d %0.2f ", inputID, goodsID, qty, cost)
	querySql := fmt.Sprintf("select id from sales_storage.goods_storage where goods_id = %d", goodsID)
	rows := h.db.ExecQuery(querySql)
	if rows == nil {
		LogError("query goods failed, when calc avg price")
		return
	}
	var id int
	for rows.Next() {
		rows.Scan(&id)
	}
	rows.Close()
	if id == 0 {
		insertSql := fmt.Sprintf("insert into sales_storage.goods_storage(goods_id , quantity, price) values (%d, %d, %f)", goodsID, qty, cost)
		err, storageID := h.db.ExecWithInsertID(insertSql)
		if err != nil {
			LogError("INSERT TO GOODS storage failed, %s", err.Error())
			return
		}
		id = int(storageID)
	} else {
		amount := float64(qty) * cost
		queryOldAmt := fmt.Sprintf("select quantity * price as amt, quantity from sales_storage.goods_storage where id = %d", id)
		rows2 := h.db.ExecQuery(queryOldAmt)
		if rows2 == nil {
			LogError("query old amt failed")
			return
		}
		var oldAmt float64
		var oldQty float64
		for rows2.Next() {
			rows2.Scan(&oldAmt, &oldQty)
		}
		rows2.Close()
		var avgPrice float64
		if float64(qty)+oldQty > 0 {
			avgPrice = (amount + oldAmt) / (float64(qty) + oldQty)
		} else {
			avgPrice = 0.0
		}

		updateSql := fmt.Sprintf("update sales_storage.goods_storage set quantity = quantity + %d, price = %f where goods_id = %d", qty, avgPrice, goodsID)
		LogInfo(updateSql)
		if err := h.db.Exec(updateSql); err != nil {
			LogError("load goods to storage table failed, %s", err.Error())
		}
	}

	if err2 := h.db.Exec("update sales_storage.goods_input set storage_id = ? where id =  ?", id, inputID); err2 != nil {
		LogError("update storage id failed, %s", err2.Error())
		return
	}
}

/*
/api/querygoods
request:
{
	"barcodeOrName" : "测试商品名字1"
}
response :
{
    "code": 0,
    "message": "success",
    "Data": [
        {
            "barcode": "1234567890",
            "goodsname": "测试商品名字1",
            "quantity": 2448,
            "cost": 102.588251,
            "sale_price": 123.55
        }
    ]
}
*/

func (h *WebHandlers) HandleQueryGoods(resp http.ResponseWriter, req *http.Request) {
	LogInfo("handle query goods request")
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		SendResponse(resp, -1, "read request failed")
		return
	}
	reqData := make(map[string]string)
	if err2 := json.Unmarshal(b, &reqData); err2 != nil {
		SendResponse(resp, -2, "parse request failed")
		return
	}
	barCodeOrName := reqData["barcodeOrName"]
	var sql string
	if h.isBarcode(barCodeOrName) {
		sql = fmt.Sprintf("select a.id, a.goods_name, a.bar_code, a.sales_price, b.quantity, b.price as cost  from sales_storage.goods_info a left join sales_storage.goods_storage b on a.id = b.goods_id  where bar_code = '%s'", barCodeOrName)
	} else {
		sql = fmt.Sprintf("select a.id, a.goods_name, a.bar_code, a.sales_price, b.quantity, b.price as cost  from sales_storage.goods_info a left join sales_storage.goods_storage b on a.id = b.goods_id  where goods_name = '%s'", barCodeOrName)
	}
	LogInfo("query sql %s", sql)
	rows := h.db.ExecQuery(sql)
	if rows == nil {
		LogError("get goods info failed")
		SendResponse(resp, -3, "get goods info failed")
		return
	}
	defer rows.Close()
	goods := &QueryGoodsResponse{
		Code:    0,
		Message: "success",
		Data:    []*GoodsInfo{},
	}
	cnt := 0
	for rows.Next() {
		cnt++
		var id int
		var goodsName, barCode string
		var salesPrice, cost, qty float64
		rows.Scan(&id, &goodsName, &barCode, &salesPrice, &qty, &cost)
		goods.Data = append(goods.Data, &GoodsInfo{
			BarCode:    barCode,
			GoodsName:  goodsName,
			Quantity:   int(qty),
			Cost:       cost,
			SalesPrice: salesPrice,
		})
	}
	if cnt == 0 {
		LogError("no goods found")
		SendResponse(resp, -5, "no goods found")
		return
	}
	b2, err2 := json.Marshal(goods)
	if err2 != nil {
		LogError("marshal json failed, %s", err2.Error())
		SendResponse(resp, -4, "marshal goods info failed")
		return
	}
	resp.Write(b2)
}

func (h *WebHandlers) isBarcode(code string) bool {
	for _, c := range code {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

/*
/api/sellgoods
request:
{
	"barcode" : "1234567890",
	"goodsname" : "测试商品名字1",
	"quantity" : 50,
	"sale_price" : 150.22,
	"location" : 1
}
response :
{
	"code" : 0 // 0 ok, or error
	"message" : "success",
}
*/
func (h *WebHandlers) HandleSellGoods(resp http.ResponseWriter, req *http.Request) {
	LogInfo("handle sell goods request")
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		SendResponse(resp, -1, "read request failed")
		return
	}
	reqData := make(map[string]interface{})
	if err2 := json.Unmarshal(b, &reqData); err2 != nil {
		SendResponse(resp, -2, "parse request failed")
		return
	}
	barCode := reqData["barcode"].(string)
	goodsName := reqData["goodsname"].(string)
	location := reqData["location"].(float64)
	qty := reqData["quantity"].(float64)
	salesPrice := reqData["sale_price"].(float64)
	LogInfo("SELL goods %s [%s] for price %0.2f quantity %0.2f", goodsName, barCode, salesPrice, qty)

	rows := h.db.ExecQuery("select a.quantity, a.goods_id, a.id from sales_storage.goods_storage a left join sales_storage.goods_info b on a.goods_id = b.id where b.bar_code = ? ", barCode)
	if rows == nil {
		SendResponse(resp, -3, "DB process failed")
		return
	}
	defer rows.Close()
	goodsQty := 0.0
	goodsID := 0
	storageID := 0
	for rows.Next() {
		rows.Scan(&goodsQty, &goodsID, &storageID)
	}
	if goodsQty < 1.0 {
		SendResponse(resp, -4, "DB process failed")
		return
	}

	if goodsQty < qty {
		SendResponse(resp, -7, "quantity is not enough")
		return
	}

	h.db.BeginTrans()
	err3 := h.db.Exec("update  sales_storage.goods_storage a left join sales_storage.goods_info b on a.goods_id = b.id set a.quantity = a.quantity - ? where b.bar_code = ?", qty, barCode)
	if err3 != nil {
		h.db.Rollback()
		SendResponse(resp, -5, "DB process failed")
		return
	}
	err4 := h.db.Exec("insert into sales_storage.goods_sales_log(goods_id, sales_location, sales_price, sales_quantity, storage_id) values (?,?,?,?,?)",
		goodsID, location, salesPrice, qty, storageID)
	if err4 != nil {
		h.db.Rollback()
		SendResponse(resp, -6, "DB process failed")
		return
	}
	h.db.Commit()
	SendResponse(resp, 0, "success")
}

/*
/api/querylocation
request:
{}
response :
{
    "code": 0,
    "message": "success",
    "Data": [
        {
            "id": 2,
            "name": "仓库1",
            "address": "",
            "updatetime": "2021-03-25 19:59:39"
        },
        {
            "id": 1,
            "name": "门店1",
            "address": "",
            "updatetime": "2021-03-24 21:05:29"
        }
    ]
}
*/
func (h *WebHandlers) HandleQueryLocation(resp http.ResponseWriter, req *http.Request) {
	rows := h.db.ExecQuery("select id, name, ifnull(address, ''), ifnull(update_time, '') from sales_storage.goods_location order by id desc")
	if rows == nil {
		SendResponse(resp, -1, "DB process failed")
		return
	}
	defer rows.Close()
	respData := &QueryLocationResponse{
		Message: "success",
		Data:    []*LocationInfo{},
	}
	for rows.Next() {
		var id int
		var name, address, updateTime string
		rows.Scan(&id, &name, &address, &updateTime)
		LogInfo("got updatetime %s", updateTime)
		info := &LocationInfo{
			ID:         id,
			Name:       name,
			Address:    address,
			UpdateTime: updateTime,
		}
		respData.Data = append(respData.Data, info)
	}
	b, err := json.Marshal(&respData)
	if err != nil {
		SendResponse(resp, -2, "DB process failed")
		return
	}
	resp.Write(b)
}
