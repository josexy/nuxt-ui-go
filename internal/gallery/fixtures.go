package gallery

import nui "github.com/josexy/nuxt-ui-go"

var propsOptions = []nui.Option[string]{{Value: "alpha", Label: "Alpha", Description: "Design"}, {Value: "beta", Label: "Beta", Description: "Engineering"}, {Value: "gamma", Label: "Gamma", Description: "Support"}}
var projectMembers = []nui.ListboxItem{{ID: "lin", Label: "Lin Chen", Group: "Design"}, {ID: "ada", Label: "Ada Lovelace", Group: "Engineering"}, {ID: "grace", Label: "Grace Hopper", Group: "Engineering"}, {ID: "guest", Label: "Guest account", Group: "External", Disabled: true}}
