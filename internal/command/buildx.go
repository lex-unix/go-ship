package command

import (
	"fmt"
	"strings"
)

func BuildImage(
	img string,
	dockerfile string,
	platform string,
	secrets map[string]string,
	buildArgs map[string]string,
) string {
	var sb strings.Builder
	sb.WriteString("docker buildx build --push --builder faino-hybrid -t ")
	sb.WriteString(img)
	sb.WriteString(fmt.Sprintf(" --platform %s", platform))
	for k := range secrets {
		sb.WriteString(fmt.Sprintf(" --secret id=%s", k))
	}
	for k, v := range buildArgs {
		sb.WriteString(fmt.Sprintf(" --build-arg %s=%q", k, v))
	}
	sb.WriteString(" ")
	sb.WriteString(dockerfile)

	return sb.String()
}
func ListBuilders(builder string) string {
	return fmt.Sprintf("docker buildx ls")
}

func CreateBuilder(builder string, driver string, platform string) string {
	return fmt.Sprintf("docker buildx create --bootstrap --platform %s --name %s --driver %s", platform, builder, driver)
}
