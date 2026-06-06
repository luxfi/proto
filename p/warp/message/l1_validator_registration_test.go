// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package message_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pcodectest"
	"github.com/luxfi/proto/p/warp/message"
)

func TestL1ValidatorRegistration(t *testing.T) {
	c := pcodectest.NewMessageCodec()
	booleans := []bool{true, false}
	for _, registered := range booleans {
		t.Run(strconv.FormatBool(registered), func(t *testing.T) {
			require := require.New(t)

			msg2, err := message.NewL1ValidatorRegistration(
				c,
				ids.GenerateTestID(),
				registered,
			)
			require.NoError(err)

			parsed, err := message.ParseL1ValidatorRegistration(c, msg2.Bytes())
			require.NoError(err)
			require.Equal(msg2, parsed)
		})
	}
}
