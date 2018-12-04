package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"ttsx/models"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

type CartController struct {
	beego.Controller
}
//添加购物车
func (this*CartController)Handleraddcart()  {
	goodsId,err1 := this.GetInt("goodsId")
	count,err2 := this.GetInt("count")
	re,err:=this.GetInt("res_code")
	resp := make(map[string]interface{})
	defer this.ServeJSON()
	if err1 != nil || err2 != nil{
		resp["res"]=1
		resp["errmsg"] = "获取数据信息错误"
		this.Data["json"] = resp
		return
	}
	//beego.Info(goodsId,count)
	userName:=this.GetSession("username")
	if userName == nil{
		resp["res"]=4
		resp["errmsg"] = "用户未登录，请提前登录"
		this.Data["json"] = resp
		return
	}
	o:=orm.NewOrm()
	var user models.User
	user.Name=userName.(string)
	//向后台数据库插入数据用read进行读取，向前端展示数据是quary
	o.Read(&user,"Name")
//	向数据库存储数据，因为数据包含三个要素，商品id，用户信息，商品数量，所以使用redis中hash来存储
    conn,err:=redis.Dial("tcp","192.168.42.188:6379")  //建立连接

	if err!=nil {
		resp["res"]=3
		resp["errmsg"] = "连接redis失败"
		this.Data["json"] = resp
		return
	}
	//县获取原来的数据

	res,err:=conn.Do("hget","Cart_"+strconv.Itoa(user.Id),goodsId)
	precount,_:=redis.Int(res,err)

	//传递数据到redis表
	if re==10 {
		conn.Do("hset","Cart_"+strconv.Itoa(user.Id),goodsId,count)
		re,err:=conn.Do("hlen","Cart_"+strconv.Itoa(user.Id))
		cartcount,_:=redis.Int(re,err)
		resp["res"] = 5
		resp["count"] = cartcount
		beego.Info(cartcount)
		this.Data["json"] = resp
		this.ServeJSON()
	}else {
		conn.Do("hset","Cart_"+strconv.Itoa(user.Id),goodsId,precount+count)
		re,err:=conn.Do("hlen","Cart_"+strconv.Itoa(user.Id))
		cartcount,_:=redis.Int(re,err)

	resp["res"] = 5
	resp["count"] = cartcount
	beego.Info(cartcount)
	this.Data["json"] = resp
	this.ServeJSON()
	}
	//beego.Info(resp)
}
//封装购物车数目显示函数
func ShowCartcount(this*beego.Controller)int  {
	userName:=this.GetSession("username")
	if userName==nil {
		return 0
	}
	o:=orm.NewOrm()
	var user models.User
	user.Name=userName.(string)
	o.Read(&user,"Name")
    conn,err1:=redis.Dial("tcp","192.168.42.188:6379")
	if err1!=nil {
		beego.Info("链接redis数据库失败")
	}
	resp,err:=conn.Do("hgetall","Cart_"+strconv.Itoa(user.Id))
	if err!=nil {
		beego.Error("数据获取失败")
	}
	goodsmap,_:=redis.IntMap(resp,err)
	var sumcount =0
	//var goods = make([]map[string]interface{},0)
	for _,count:=range goodsmap{
		sumcount+=count
	}
	return sumcount
}
//显示订单页面
func (this*CartController)Showcart()  {
	//获取数据,因为是从redis获取数据，所以线连接redis
	conn,err:=redis.Dial("tcp","192.168.42.188:6379")
	if err!=nil {
		beego.Error("redis连接失败")
	}
	userName:=this.GetSession("username")
	//校验数据
	//处理数据
	o:=orm.NewOrm()
	var user models.User
	user.Name=userName.(string)
	o.Read(&user,"Name")
//	利用已经获取的用户ID 获取redis购物车数据
	resp,err:=conn.Do("hgetall","Cart_"+strconv.Itoa(user.Id))
	goodsmap,_:=redis.IntMap(resp,err)
	//beego.Info(goodsmap)
//   因为获得的数据是key value类型，key为string类型，value是一个结构体，所以定义一个容器，用于储存数据
	var goods = make([]map[string]interface{},0)
	for goodsId,count:=range goodsmap{
	//	定义一个临时容器用于接收临时接收数据
	temp:=make(map[string]interface{})
	id,_:=strconv.Atoi(goodsId)
	var goodsSku models.GoodsSKU
	goodsSku.Id=id
	o.Read(&goodsSku)
	temp["goodssku"]=goodsSku
	temp["count"]=count
	temp["sumprice"]=count*goodsSku.Price
	goods=append(goods,temp)
	}

//	返回数据
    this.Data["goods"]=goods
    //beego.Info(goods)
    this.TplName="cart.html"

}
//删除订单页商品
func (this*CartController)Handledel()  {
//	获取商品ID 从redis中获取，建立连接
goodsId,err:=this.GetInt("goodsId")
//判断数据是否有错，如果有错，返回错误信息给ajax
//这一段是创建一个容器，接受ajax传过来的res，如果信息不是制定的值，就返回错误信息
  resp:=make(map[string]interface{})
 /* resp["res"]=5
  resp["errmsg"]="ok"*/
  defer this.ServeJSONP()
	if err!=nil {
		resp["res"]=1
		resp["errmsg"] = "无效商品id"
		this.Data["json"] = resp
		return
	}
//	处理数据 处理数据需要删除redis里面hash存储的三个值，一个是用户ID，一个是商品信息，一个是数量
//所以获取用户信息

  userName:=this.GetSession("username")
  o:=orm.NewOrm()
  var user models.User
  user.Name=userName.(string)
  o.Read(&user,"Name")
conn,_:=redis.Dial("tcp","192.168.42.188:6379")

	_,err=conn.Do("hdel","Cart_"+strconv.Itoa(user.Id),goodsId)
	if err!=nil {
		resp["res"] = 3
		resp["errmsg"] = "删除商品失败"
		this.Data["json"] = resp
		return
	}
	resp["res"] = 5
/*	resp["errmsg"] = "ok"*/
	this.Data["json"] = resp
	this.ServeJSON()

}