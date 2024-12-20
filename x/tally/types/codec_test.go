package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeFilterInput(t *testing.T) {
	tests := []struct {
		name    string
		hexStr  string
		want    string
		wantErr error
	}{
		{
			name:    "happy path",
			hexStr:  "01000000000000000d242e726573756c742e74657874",
			want:    "$.result.text",
			wantErr: nil,
		},
		{
			name:    "happy path",
			hexStr:  "01000000000000000b726573756C742E74657874",
			want:    "result.text",
			wantErr: nil,
		},
		{
			name:    "a long json path",
			hexStr:  "010000000000000097746869732e69732e616e2e756e7265616c6973746963616c6c792e6c6f6e672e646174612e706174682e746861742e73686f756c642e6e657665722e68617070656e2e696e2e7265616c2e776f726c642e6275742e69742e69732e612e676f6f642e746573742e636173652e746f2e636865636b2e7468652e66756e6374696f6e616c6974792e6f662e756e7061636b2e6d6574686f64",
			want:    "this.is.an.unrealistically.long.data.path.that.should.never.happen.in.real.world.but.it.is.a.good.test.case.to.check.the.functionality.of.unpack.method",
			wantErr: nil,
		},
		{
			name:    "invalid len - Actual len(data) bigger than encoded len",
			hexStr:  "010000000000000009242e726573756c742e74657874",
			want:    "",
			wantErr: ErrInvalidPathLen,
		},
		{
			name:    "invalid len - Actual len small",
			hexStr:  "010000000000000019242e726573756c742e74657874",
			want:    "",
			wantErr: ErrInvalidPathLen,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := hex.DecodeString(tt.hexStr)
			require.NoError(t, err)

			filter, err := NewFilterMode(b, 1, 1)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, filter.dataPath)
		})
	}
}
