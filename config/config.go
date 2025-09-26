package config

import "gopkg.in/yaml.v3"

type Config struct {
	Basic Basic `yaml:"basic"`
	Cache Cache `yaml:"cache"`
	Log   Log   `yaml:"log"`
	Db    Db    `yaml:"db"`
}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	type plain Config
	if err := value.Decode((*plain)(c)); err != nil {
		return err
	}

	// 根据DbEngine从对应节点获取DB配置
	switch c.Basic.DbEngine {
	case "mysql":
		if mysqlNode := findChildNode(value, "mysql"); mysqlNode != nil {
			return mysqlNode.Decode(&c.Db)
		}
	case "postgres":
		if pgNode := findChildNode(value, "postgres"); pgNode != nil {
			return pgNode.Decode(&c.Db)
		}
	case "sqlite":
		if sqliteNode := findChildNode(value, "sqlite"); sqliteNode != nil {
			return sqliteNode.Decode(&c.Db)
		}
	case "mssql":
		if mssqlNode := findChildNode(value, "mssql"); mssqlNode != nil {
			return mssqlNode.Decode(&c.Db)
		}
	}
	return nil
}

func findChildNode(node *yaml.Node, name string) *yaml.Node {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == name {
			return node.Content[i+1]
		}
	}
	return nil
}
