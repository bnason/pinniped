// This go.mod file is generated by ./hack/codegen.sh.
module go.pinniped.dev/generated/1.22/client

go 1.13

require (
	go.pinniped.dev/generated/1.22/apis v0.0.0
	k8s.io/apimachinery v0.22.13
	k8s.io/client-go v0.22.13
	k8s.io/kube-openapi v0.0.0-20211109043538-20434351676c
)

replace go.pinniped.dev/generated/1.22/apis => ../apis
