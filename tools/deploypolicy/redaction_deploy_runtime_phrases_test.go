package deploypolicy

func deployRuntimeRequiredPhrases() []string {
	return []string{
		"profile-thumbnails/uploads",
		"profile_thumbnail_url",
		"(.devices | type) == \"array\"",
		"(.devices[0].device_id // \"\")",
		"if [ -n \"$device_id\" ]",
		"printf '%s' \"$image_uri\" > \"$RUNNER_TEMP/riido-image-uri\"",
		"printf '%s' \"$next_task_definition\" > \"$RUNNER_TEMP/riido-task-definition-arn\"",
		"printf '%s' \"$container_port\" > \"$RUNNER_TEMP/riido-container-port\"",
		"umask 077",
		"chmod 600 \"$current_json\"",
		"chmod 600 \"$next_json\"",
		"Cleanup live handoff files",
		"if: always()",
		"rm -f \\",
		"\"$RUNNER_TEMP/riido-image-uri\"",
		"\"$RUNNER_TEMP/riido-task-definition-arn\"",
		"\"$RUNNER_TEMP/riido-container-port\"",
		"current_json=\"$RUNNER_TEMP/task-definition.current.json\"",
		"next_json=\"$RUNNER_TEMP/task-definition.next.json\"",
		"appspec_json=\"$RUNNER_TEMP/codedeploy-appspec.json\"",
		"deployment_json=\"$RUNNER_TEMP/codedeploy-deployment.json\"",
		"revisionType: \"AppSpecContent\"",
		"aws deploy create-deployment",
		"wait_deployment_id=\"$(cat \"$deployment_id_file\")\"",
		"echo \"::add-mask::$wait_deployment_id\"",
		"aws deploy wait deployment-successful",
	}
}
