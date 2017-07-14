package config

import (
	"fmt"
	"bytes"
)

type ForLoad struct{
	Groups []*GroupConfig
	Pool PoolConfig
}

func (c *ForLoad) Show() string {
	info := new(bytes.Buffer)

	for _, groupc := range c.Groups {
		info.WriteString(fmt.Sprintf("group #%d-%s#:{\r\n", groupc.Id, groupc.Name))
		for _, bunchc := range groupc.Bunches {
			backup := ""
			if len(bunchc.Backups) > 0 {
				backup = bunchc.Backups[0]

				for i:=1;i<len(bunchc.Backups);i++ {
					backup += "," + bunchc.Backups[i]
				}
			}

			sinfo := fmt.Sprintf("bunch #%d#:[primary:(%s), backup:(%s)]", bunchc.Id, bunchc.Primary, backup)
			info.WriteString(fmt.Sprintf("%s\r\n", sinfo))
		}
		info.WriteString("}\r\n")
	}

	return info.String()
}