### /api/inputgoods  商品入库

```
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
```

### /api/querygoods 查询商品

```
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
```

### /api/sellgoods  销售商品

```
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
```

### /api/querylocation 查询门店列表
```
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
```

