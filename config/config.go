package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

var CNF *config = &config{}

/*
Priority:
  - env
  - YAML
  - default

Notes:
 1. make sure to give a default value as possible
 2. the default values are dev-environment oriented
 3. preferably use YAML unless it's a security concern
*/
type config struct {
	Env     string `env:"ENV" yaml:"Env" env-default:"dev" env-description:"dev, stg, or prod"`
	Service struct {
		Name  string `env:"SERVICE__NAME" yaml:"Name" env-default:"client"`
		Title string `env:"SERVICE__TITLE" yaml:"Title" env-default:"client Business (Customer) Domain"`
		Root  string `env:"SERVICE__ROOT" yaml:"Root" env-default:"."`
	} `yaml:"Service"`
	App struct{} `yaml:"App"` // for business-specific only
	Db  struct {
		DSN      string `env:"DB_DSN" yaml:"DB_DSN" env-default:""`
		Host     string `env:"DB__HOST" yaml:"Host" env-default:"localhost"`
		Port     int    `env:"DB__PORT" yaml:"Port" env-default:"54322"`
		User     string `env:"DB__USER" yaml:"User" env-default:"clientadmin"`
		Password string `env:"DB__PASSWORD" yaml:"Password" env-default:"clientpassword"`
		Name     string `env:"DB__NAME" yaml:"Name" env-default:"clientdb"`
	} `yaml:"Db"`
	Temporal struct {
		HostPort  string `env:"TEMPORAL__HOSTPORT" yaml:"HostPort" env-default:""`
		Namespace string `env:"TEMPORAL__NAMESPACE" yaml:"Namespace" env-default:""`
	} `yaml:"Temporal"`
	// Keycloak owns login/credentials/sessions — it is NOT this domain (design §0, §4.3).
	// This domain only holds the `sub` reference; these values are infra config used by
	// the identity-provider anti-corruption ports to provision users/organizations.
	Keycloak struct {
		BaseURL string `env:"KEYCLOAK__BASE_URL" yaml:"BaseURL" env-default:"http://localhost:8081"`
		Realm   string `env:"KEYCLOAK__REALM" yaml:"Realm" env-default:"client"`
		// AdminToken is a bearer token for the Keycloak Admin API. In a full build this
		// is minted via a client-credentials grant; here it is injected as config so the
		// adapter stays a thin, faithful mirror (design of api-adapter-standard §1).
		AdminToken string `env:"KEYCLOAK__ADMIN_TOKEN" yaml:"AdminToken" env-default:""`
	} `yaml:"Keycloak"`
	Srv struct {
		Host string `env:"SERVER__HOST" yaml:"Host" env-default:""`
		Port string `env:"SERVER__PORT" yaml:"Port" env-default:"9000"`
	} `yaml:"Srv"`
}

func Load() {
	cleanenv.ReadEnv(CNF)
	_ = cleanenv.ReadConfig(CNF.Service.Root+"/config/.env", CNF)

	err := cleanenv.ReadConfig(CNF.Service.Root+"/config/"+CNF.Env+".yml", CNF)
	handleConfigErr(err)

	logger := setSlog()
	slog.SetDefault(logger)
	logger.Info("configuration loaded", "CNF", CNF)
}

func setSlog() *slog.Logger {
	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})
	return slog.New(handler)
}

func handleConfigErr(err error) {
	if err != nil {
		if strings.Contains(err.Error(), "config file parsing error: EOF") {
			return
		}
		panic(err)
	}
}
