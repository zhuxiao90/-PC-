package controllers

import ("github.com/astaxie/beego"
        "ttsx/models"
        "github.com/astaxie/beego/orm"
	    "regexp"
	    "github.com/astaxie/beego/utils"
	"strconv"
)


type UserController struct {
	beego.Controller
}

func (this*UserController)ShowRegister()  {
	this.TplName="register.html"
}

func (this*UserController)HandleRegister()  {
//	接收数据
userName:=this.GetString("user_name")
pwd:=this.GetString("pwd")
cpwd := this.GetString("cpwd")
email:=this.GetString("email")

//校验数据
	if userName==""||pwd==""||email==""||cpwd=="" {
		this.Data["errmsg"]="不能为空"
		this.TplName="register.html"
		return
	}

	reg,_ :=regexp.Compile("^[\\w-]+(\\.[\\w-]+)*@[\\w-]+(\\.[\\w-]+)+$")
    res:=reg.FindString(email)
	if res=="" {
		this.Data["errmsg"]="邮箱格式错误"
		this.TplName="register.html"
		return
	}
	if pwd != cpwd {
		this.Data["errmsg"] = "两次密码输入不正确，请重新注册！"
		this.TplName = "register.html"
		return
	}
//处理数据
//获取orm对象
o:=orm.NewOrm()

//  获取插入对象
var user models.User
//给插入对象赋值
user.Name=userName
user.PassWord=pwd
user.Email=email

//插入数据
  _,err:=o.Insert(&user)
	if err!=nil {
		this.Data["errmsg"]="插入数据失败"
		this.TplName="register.html"
		return
	}
	config:=`{"username":"zfawsdsb@163.com","password":"zhuzhu530611","host":"smtp.163.com","port":25}`
	temail:=utils.NewEMail(config)
	temail.To = []string{email}
	temail.From = "zfawsdsb@163.com"
	//temail.HTML = "复制该连接到浏览器中激活：127.0.0.1:8088/active?id="+strconv.Itoa(user.Id)
	temail.HTML="<a href=\"http://127.0.0.1:8080/active?id="+strconv.Itoa(user.Id)+"\">点击该链接，天天生鲜用户激活</a>"
	err = temail.Send()
	if err != nil{
		this.Data["errmsg"] = "发送激活邮件失败，请重新注册！"
		this.TplName = "register.html"
		return
	}

//返回数据
	this.Ctx.WriteString("注册成功,激活邮箱去吧")


}
func (this*UserController)HandleActive()  {
//	接收数据
    id,err:=this.GetInt("id")

//验证数据
	if err!=nil {
		this.Data["errmsg"] = "激活失败，请重新注册！"
		this.TplName = "register.html"
		return
	}
//处理数据
//  设置orm随想
o:=orm.NewOrm()
var user models.User
//赋值确定查询的表名
user.Id=id
//查询
  err=o.Read(&user)
	if err!=nil {
		this.Data["errmsg"] = "激活失败，请重新注册！"
		this.TplName = "register.html"
		return
	}
//	给需要更新的字段赋值
user.Active=true
// 更新数据
  _,err=o.Update(&user)
	if err!=nil {
		this.Data["errmsg"] = "激活失败，请重新注册！"
		this.TplName = "register.html"
		return
	}
//返回数据
	this.Redirect("/login",302)

}
func (this*UserController)ShowLogin()  {
	userName:=this.Ctx.GetCookie("username")
	if userName=="" {
		this.Data["username"]=""
		this.Data["checked"]=""
	}else {
		this.Data["username"]=userName
		this.Data["checked"]="checked"
	}
	this.TplName="login.html"
}
func (this*UserController)HandleLogin()  {
//	相同四步
    userName:=this.GetString("username")
    pwd :=this.GetString("pwd")
	if userName==""||pwd=="" {
		this.Data["errmsg"]="用户不存在，清重新登陆"
		this.TplName="login.html"
		return
	}
//	处理数据
    o:=orm.NewOrm()
    var uer models.User
    uer.Name=userName
//	查询数据，确定是哪个用胡或者有没有这个用户
    err:=o.Read(&uer,"Name")
	if err!=nil {
		this.Data["errmsg"]="用户不存在，清重新登陆"
		this.TplName="login.html"
		return
	}
	if uer.PassWord!=pwd {
		this.Data["errmsg"]="mima不存在，清重新登陆"
		this.TplName="login.html"
		return
	}
	if !uer.Active {
		this.Data["errmsg"]="用户未激活，清重新登陆"
		this.TplName="login.html"
		return
	}
	check:=this.GetString("check")
	if check=="on" {
		this.Ctx.SetCookie("username",userName,3600)
	}else {
		this.Ctx.SetCookie("username",userName,-1)
	}
	this.SetSession("username",userName)
	this.Redirect("/index",302)

}