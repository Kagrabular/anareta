package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1alpha1 "github.com/Kagrabular/ANARETA/api/v1alpha1"
)

// WebhookPayload represents a minimal GitHub PR event payload
type WebhookPayload struct {
	Action      string `json:"action"`
	PullRequest struct {
		Head struct {
			Ref  string `json:"ref"`
			Repo struct {
				CloneURL string `json:"clone_url"`
			} `json:"repo"`
		} `json:"head"`
	} `json:"pull_request"`
}

// MakeWebhookHandler returns an HTTP handler for processing GitHub webhooks
func MakeWebhookHandler(k8sClient client.Client, namespace string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		eventType := r.Header.Get("X-GitHub-Event")
		if eventType != "pull_request" {
			fmt.Fprintf(w, "Event %s ignored", eventType)
			return
		}

		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read payload", http.StatusBadRequest)
			return
		}

		var prEvent WebhookPayload
		if err := json.Unmarshal(payload, &prEvent); err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}

		// clean branch name for resource naming
		branchSafe := strings.ReplaceAll(prEvent.PullRequest.Head.Ref, "/", "-")
		crName := branchSafe
		ctx := context.Background()

		switch prEvent.Action {
		case "opened", "reopened", "synchronize":
			dev := &appv1alpha1.DevEnv{
				ObjectMeta: metav1.ObjectMeta{
					Name:      crName,
					Namespace: namespace,
				},
				Spec: appv1alpha1.DevEnvSpec{
					RepoURL: prEvent.PullRequest.Head.Repo.CloneURL,
					Branch:  prEvent.PullRequest.Head.Ref,
					TTL:     metav1.Duration{Duration: 24 * time.Hour},
				},
			}
			if err := k8sClient.Create(ctx, dev); err != nil && !client.IgnoreAlreadyExists(err) {
				http.Error(w, fmt.Sprintf("Failed to create DevEnv: %v", err), http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "DevEnv %s created\n", crName)

		case "closed":
			dev := &appv1alpha1.DevEnv{
				ObjectMeta: metav1.ObjectMeta{
					Name:      crName,
					Namespace: namespace,
				},
			}
			if err := k8sClient.Delete(ctx, dev); err != nil && !client.IgnoreNotFound(err) {
				http.Error(w, fmt.Sprintf("Failed to delete DevEnv: %v", err), http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "DevEnv %s deleted\n", crName)

		default:
			fmt.Fprintf(w, "Action %s ignored\n", prEvent.Action)
		}
	}
}
