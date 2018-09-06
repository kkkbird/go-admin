package controller

import (
	"bytes"
	"encoding/json"
	"goAdmin/modules/connections"
	"goAdmin/modules/auth"
	"goAdmin/plugins/admin/models"
	"goAdmin/context"
	"goAdmin/modules/menu"
	"goAdmin/template"
	"goAdmin/template/types"
	"net/http"
)

// 显示菜单
func ShowMenu(ctx *context.Context) {
	defer GlobalDeferHandler(ctx)

	path := ctx.Path()
	user := ctx.UserValue["user"].(auth.User)

	menu.GlobalMenu.SetActiveClass(path)

	editUrl := Config.ADMIN_PREFIX + "/menu/edit/show"
	deleteUrl := Config.ADMIN_PREFIX + "/menu/delete"
	orderUrl := Config.ADMIN_PREFIX + "/menu/delete"

	tree := template.Get(Config.THEME).Tree().SetTree((*menu.GlobalMenu).GlobalMenuList).
		SetEditUrl(editUrl).SetDeleteUrl(deleteUrl).SetOrderUrl(orderUrl).GetContent()
	header := template.Get(Config.THEME).Tree().GetTreeHeader()
	box := template.Get(Config.THEME).Box().SetHeader(header).SetBody(tree).GetContent()
	col1 := template.Get(Config.THEME).Col().SetSize(map[string]string{"md": "6"}).SetContent(box).GetContent()
	newForm := template.Get(Config.THEME).Form().SetPrefix(Config.ADMIN_PREFIX).SetUrl(Config.ADMIN_PREFIX + "/menu/new").SetInfoUrl(Config.ADMIN_PREFIX + "/menu").SetTitle("New").
		SetContent(models.GetNewFormList(models.GlobalTableList["menu"].Form.FormList)).GetContent()
	col2 := template.Get(Config.THEME).Col().SetSize(map[string]string{"md": "6"}).SetContent(newForm).GetContent()
	row := template.Get(Config.THEME).Row().SetContent(col1 + col2).GetContent()

	tmpl, tmplName := template.Get("adminlte").GetTemplate(ctx.Request.Header.Get("X-PJAX") == "true")

	menu.GlobalMenu.SetActiveClass(path)

	ctx.Response.Header.Add("Content-Type", "text/html; charset=utf-8")

	buf := new(bytes.Buffer)

	tmpl.ExecuteTemplate(buf, tmplName, types.Page{
		User: user,
		Menu: *menu.GlobalMenu,
		System: types.SystemInfo{
			"0.0.1",
		},
		Panel: types.Panel{
			Content:     row,
			Description: "菜单管理",
			Title:       "菜单管理",
		},
		AssertRootUrl: Config.ADMIN_PREFIX,
	})

	ctx.WriteString(buf.String())
}

// 显示编辑菜单
func ShowEditMenu(ctx *context.Context) {
	id := ctx.Request.URL.Query().Get("id")
	prefix := "menu"

	formData, title, description := models.GlobalTableList[prefix].GetDataFromDatabaseWithId(prefix, id)

	tmpl, tmplName := template.Get("adminlte").GetTemplate(ctx.Request.Header.Get("X-PJAX") == "true")

	path := ctx.Path()
	menu.GlobalMenu.SetActiveClass(path)

	ctx.Response.Header.Add("Content-Type", "text/html; charset=utf-8")
	user := ctx.UserValue["user"].(auth.User)

	buf := new(bytes.Buffer)
	tmpl.ExecuteTemplate(buf, tmplName, types.Page{
		User: user,
		Menu: *menu.GlobalMenu,
		System: types.SystemInfo{
			"0.0.1",
		},
		Panel: types.Panel{
			Content:     template.Get(Config.THEME).Form().
				SetContent(formData).
				SetPrefix(Config.ADMIN_PREFIX).
				SetUrl(Config.ADMIN_PREFIX + "/edit/" + prefix).
				SetToken(auth.TokenHelper.AddToken()).
				SetInfoUrl(Config.ADMIN_PREFIX + "/info/" + prefix).
				GetContent(),
			Description: description,
			Title:       title,
		},
		AssertRootUrl: Config.ADMIN_PREFIX,
	})
	ctx.WriteString(buf.String())
}

// 删除菜单
func DeleteMenu(ctx *context.Context) {
	id := ctx.Request.URL.Query().Get("id")

	buffer := new(bytes.Buffer)

	connections.GetConnection().Exec("delete from goadmin_menu where id = ?", id)

	menu.SetGlobalMenu()
	//template.MenuPanelPjax((*menu.GlobalMenu).GetEditMenuList(), (*menu.GlobalMenu).GlobalMenuOption, buffer)

	ctx.WriteString(buffer.String())
	ctx.Response.Header.Add("Content-Type", "text/html; charset=utf-8")
}

// 编辑菜单
func EditMenu(ctx *context.Context) {
	defer GlobalDeferHandler(ctx)

	buffer := new(bytes.Buffer)

	id := string(ctx.Request.FormValue("id")[:])
	title := string(ctx.Request.FormValue("title")[:])
	parentId := string(ctx.Request.FormValue("parent_id")[:])
	if parentId == "" {
		parentId = "0"
	}
	icon := string(ctx.Request.FormValue("icon")[:])
	uri := string(ctx.Request.FormValue("uri")[:])

	connections.GetConnection().Exec("update goadmin_menu set title = ?, parent_id = ?, icon = ?, uri = ? where id = ?",
		title, parentId, icon, uri, id)

	menu.SetGlobalMenu()

	//template.MenuPanelPjax((*menu.GlobalMenu).GetEditMenuList(), (*menu.GlobalMenu).GlobalMenuOption, buffer)

	ctx.WriteString(buffer.String())
	ctx.Response.Header.Add("Content-Type", "text/html; charset=utf-8")
	ctx.Response.Header.Add("X-PJAX-URL", Config.ADMIN_PREFIX + "/menu")
}

// 新建菜单
func NewMenu(ctx *context.Context) {
	defer GlobalDeferHandler(ctx)

	buffer := new(bytes.Buffer)

	title := string(ctx.Request.FormValue("title")[:])
	parentId := string(ctx.Request.FormValue("parent_id")[:])
	if parentId == "" {
		parentId = "0"
	}
	icon := string(ctx.Request.FormValue("icon")[:])
	uri := string(ctx.Request.FormValue("uri")[:])

	connections.GetConnection().Exec("insert into goadmin_menu (title, parent_id, icon, uri, `order`) values (?, ?, ?, ?, ?)", title, parentId, icon, uri, (*menu.GlobalMenu).MaxOrder+1)

	(*menu.GlobalMenu).SexMaxOrder((*menu.GlobalMenu).MaxOrder + 1)
	menu.SetGlobalMenu()

	//template.MenuPanelPjax((*menu.GlobalMenu).GetEditMenuList(), (*menu.GlobalMenu).GlobalMenuOption, buffer)

	ctx.WriteString(buffer.String())
	ctx.Response.Header.Add("Content-Type", "text/html; charset=utf-8")
	ctx.Response.Header.Add("X-PJAX-URL", Config.ADMIN_PREFIX + "/menu")
}

// 修改菜单顺序
func MenuOrder(ctx *context.Context) {
	defer GlobalDeferHandler(ctx)

	var data []map[string]interface{}
	json.Unmarshal([]byte(ctx.Request.FormValue("_order")), &data)

	count := 1
	for _, v := range data {
		if child, ok := v["children"]; ok {
			connections.GetConnection().Exec("update goadmin_menu set `order` = ? where id = ?", count, v["id"])
			for _, v2 := range child.([]interface{}) {
				connections.GetConnection().Exec("update goadmin_menu set `order` = ? where id = ?", count, v2.(map[string]interface{})["id"])
				count++
			}
		} else {
			connections.GetConnection().Exec("update goadmin_menu set `order` = ? where id = ?", count, v["id"])
			count++
		}
	}
	menu.SetGlobalMenu()

	ctx.SetStatusCode(http.StatusOK)
	ctx.SetContentType("application/json")
	ctx.WriteString(`{"code":200, "msg":"ok"}`)
	return
}
