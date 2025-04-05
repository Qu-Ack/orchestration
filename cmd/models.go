package main

type WebhookPayloadCommitAuthorStruct struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type WebhookPayloadCommitStruct struct {
	CommitHash    string                           `json:"id"`
	TreeHash      string                           `json:"tree_id"`
	CommitMessage string                           `json:"message"`
	CommitUrl     string                           `json:"url"`
	Author        WebhookPayloadCommitAuthorStruct `json:"author"`
}
type WebhookPayloadRepositoryStruct struct {
	Name     string `json:"full_name"`
	Url      string `json:"url"`
	CloneUrl string `json:"clone_url"`
}

type WebhookPayloadStruct struct {
	Ref        string                         `json:"ref"`
	Repository WebhookPayloadRepositoryStruct `json:"repository"`
	Commits    []WebhookPayloadCommitStruct   `json:"commits"`
}

type WebhookPayloadBody struct {
	Payload WebhookPayloadStruct `json:"payload"`
}
