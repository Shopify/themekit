package cmdutil

import (
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/shopify"
)

type shopifyClient interface {
	GetShop() (shopify.Shop, error)
	CreateNewTheme(string, string) (shopify.Theme, error)
	GetInfo() (shopify.Theme, error)
	Themes() ([]shopify.Theme, error)
	GetAllAssets() ([]string, error)
	GetAsset(string) (shopify.Asset, error)
	UpdateAsset(shopify.Asset) error
	DeleteAsset(shopify.Asset) error
}

type config interface {
	Set(string, env.Env, ...env.Env) (*env.Env, error)
	Get(string, ...env.Env) (*env.Env, error)
	Save() error
}
