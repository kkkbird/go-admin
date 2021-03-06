package action

import (
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/template/types"
	"html/template"
)

type FieldFilterAction struct {
	BaseAction
	Prefix string
	Field  string
}

func FieldFilter(prefix, field string) *FieldFilterAction {
	return &FieldFilterAction{Prefix: prefix, Field: field}
}

func (jump *FieldFilterAction) ExtContent() template.HTML {

	options := jump.BtnData.(types.FieldOptions)

	cm := ``
	for _, obejct := range options {
		cm += `if (e.params.data.text === "` + obejct.Text + `") {
		$.pjax({url: "` + config.Get().Url("/info/"+jump.Prefix+"?"+jump.Field+"="+obejct.Value) + `&` + form.NoAnimationKey +
			`=true", container: '#pjax-container'});
	}`
	}

	return template.HTML(`<script>
$(".` + jump.BtnId + `").on("select2:select",function(e){
	` + cm + `
})
vv = ""
query = window.location.search.substring(1);
vars = query.split("&");
for (let i = 0; i < vars.length; i++) {
	pair = vars[i].split("=");
	if (pair[0] === "` + jump.Field + `") {
		vv = pair[1];
	}
}
if (vv !== "") {
	$(".` + jump.BtnId + `").val(vv).select2()
}
</script>`)
}
