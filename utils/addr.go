package utils

import (
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func ConvertAddressPrefixes(addr string, prefixes []string) []string {
	if addr == "" {
		return nil
	}
	_, bz, err := bech32.DecodeAndConvert(addr)
	if err != nil {
		return []string{addr}
	}
	convertedAddrs := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		convertedAddr, err := bech32.ConvertAndEncode(prefix, bz)
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
		for _, convertedAddr := range ConvertAddressPrefixes(addr, prefixes) {
			convertedAddrs = append(convertedAddrs, convertedAddr)
		}
	}
	return convertedAddrs
}
