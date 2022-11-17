package utils

import (
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func ConvertAddressPrefix(addr, prefix string) (string, error) {
	_, bz, err := bech32.DecodeAndConvert(addr)
	if err != nil {
		return "", err
	}
	return bech32.ConvertAndEncode(prefix, bz)
}

func ConvertAddressPrefixes(addr string, prefixes []string) []string {
	if addr == "" {
		return nil
	}
	convertedAddrs := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		convertedAddr, err := ConvertAddressPrefix(addr, prefix)
		if err != nil {
			return []string{addr}
		}
		convertedAddrs = append(convertedAddrs, convertedAddr)
	}
	return convertedAddrs
}

// basically it's `addrs.flatMap((addr) => ConvertAddressPrefixes(addr, prefixes))`
func ConvertAddressArrayPrefixes(addrs []string, prefixes []string) []string {
	convertedAddrs := []string{}
	for _, addr := range addrs {
		convertedAddrs = append(convertedAddrs, ConvertAddressPrefixes(addr, prefixes)...)
	}
	return convertedAddrs
}
