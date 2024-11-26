package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"

	"github.com/navigatore300/netangelsgo"
	log "github.com/sirupsen/logrus"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var GroupName = os.Getenv("GROUP_NAME")
var LogLevel = os.Getenv("LOG_LEVEL")

func main() {
	time.Sleep(10 * time.Second)
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName,
		&NetangelsDnsSolver{},
	)
}

type NetangelsDNSProviderConfig struct {
	// Change the two fields below according to the format of the configuration
	// to be decoded.
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.
	SecretRef   string `json:"secretName"`
	AccountName string `json:"accountName"`
	ApiKey      string `json:"apiKey"`
}

type NetangelsDnsSolver struct {
	client     netangelsgo.NetangelsClient
	kubeClient *kubernetes.Clientset
}

func (e *NetangelsDnsSolver) Name() string {
	return "netangels-dns-solver"
}
func (e *NetangelsDnsSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	log.Info("Presenting challenge for: ", ch.ResolvedFQDN)
	err := loadCredentials(ch, e)
	if err != nil {
		log.Errorf("load credentials failed(check secret configuration): %v", err)
		return err
	}

	id, txtData, fetchErr := e.client.GetRecord(ch.ResolvedFQDN, ch.Key, "TXT")
	if fetchErr == nil && id != 0 && txtData != ch.Key {
		_, err := e.client.UpdateRecord(id, ch.ResolvedFQDN, ch.Key, "TXT", 300)
		if err != nil {
			log.Errorf("presenting challenge failed: %v", err)
			return err
		}
		log.Debug("Challenge have been created with record id: ", id)
		return nil
	} else if fetchErr == nil && id != 0 && txtData == ch.Key {
		log.Debug("Challenge have been created with record id: ", id)
		return nil
	} else {
		id, err = e.client.AddRecord(ch.ResolvedFQDN, ch.Key, "TXT", 300)
		if err != nil {
			log.Errorf("presenting challenge failed: %v", err)
			return err
		} else {
			log.Debug("Challenge have been created with record id: ", id)
		}
		return nil
	}
}

func (e *NetangelsDnsSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	log.Info("Cleaning up challenge for: ", ch.ResolvedFQDN)
	err := loadCredentials(ch, e)
	if err != nil {
		log.Errorf("load credentials failed(check secret configuration): %v", err)
		return err
	}
	id, _, err := e.client.GetRecord(ch.ResolvedFQDN, ch.Key, "TXT")
	if err != nil {
		log.Errorf("error on fetching record: %v", err)
		return err
	}
	log.Debug("Record(", id, ") fetched for cleanup.")

	if err := e.client.RemoveRecord(id); err != nil {
		log.Errorf("record(%d) have not been cleaned up", id)
		return err
	} else {
		log.Debug("Record(", id, ") have been cleaned up.")
		return nil

	}

}

func (e *NetangelsDnsSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	if LogLevel == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.Info("Initializing netangels dns solver")
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		log.Errorf("init failed with error: %v", err)
		return err
	}
	e.kubeClient = cl
	return nil
}

func loadConfig(cfgJSON *extapi.JSON) (NetangelsDNSProviderConfig, error) {
	log.Debug("Loading config...")
	cfg := NetangelsDNSProviderConfig{}

	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		log.Errorf("error decoding solver config: %v", err)
		return cfg, errors.New("error decoding solver config")
	}
	log.Debug("Config loaded successfully.")
	return cfg, nil
}

func stringFromSecretData(secretData *map[string][]byte, key string) (string, error) {
	data, ok := (*secretData)[key]
	if !ok {
		log.Errorf("key %q not found in secret data", key)
		return "", errors.New("key not found in secret data")
	}
	return string(data), nil
}

func loadCredentials(ch *v1alpha1.ChallengeRequest, e *NetangelsDnsSolver) error {
	if e.client.Credentials.AccountName != "" && e.client.Credentials.ApiKey != "" {
		return nil
	}

	cfg, err := loadConfig(ch.Config)
	if err != nil {
		log.Errorf("error on loading config: %v", err)
		return err
	}
	if cfg.AccountName != "" && cfg.ApiKey != "" {
		log.Debug("Loading API credentials from config.")
		e.client = netangelsgo.CreateNetangelsClient(cfg.AccountName, cfg.ApiKey)
		return nil
	} else {
		secretName := cfg.SecretRef
		log.Debug("Loading API credentials from secret: ", secretName)
		sec, err := e.kubeClient.CoreV1().Secrets(ch.ResourceNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			log.Errorf("error on loading secret from kubernetes api: %v", err)
			return err
		}

		accountName, err := stringFromSecretData(&sec.Data, "account-name")
		if err != nil {
			log.Errorf("error on reading secret: %v", err)
			return err
		}
		apiKey, err := stringFromSecretData(&sec.Data, "api-key")
		if err != nil {
			log.Errorf("error on reading secret: %v", err)
			return err
		}

		e.client = netangelsgo.CreateNetangelsClient(accountName, apiKey)
		return nil
	}
}
