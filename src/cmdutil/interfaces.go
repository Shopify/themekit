package cmdutil

import (
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/shopify"
)

type shopifyClient interface {
	GetShop() (shopify.Shop, error)
	CreateNewTheme(string) (shopify.Theme, error)
	GetInfo() (shopify.Theme, error)
	PublishTheme() error
	Themes() ([]shopify.Theme, error)
	GetAllAssets() ([]shopify.Asset, error)
	GetAsset(string) (shopify.Asset, error)
	UpdateAsset(shopify.Asset, string) error
	DeleteAsset(shopify.Asset) error
}

type config interface {
	Set(string, env.Env, ...env.Env) (*env.Env, error)
	Get(string, ...env.Env) (*env.Env, error)
	Save() error
}
