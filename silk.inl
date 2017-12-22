#include <stdio.h>
#include <stdlib.h>
#include "SKP_Silk_SDK_API.h"

void *new_decoder()
{
	SKP_int16 ret = 0;
	SKP_int32 size = 0;
	void *d = NULL;

    SKP_Silk_SDK_Get_Decoder_Size( &size );
    d = malloc( size );
    ret = SKP_Silk_SDK_InitDecoder( d );
    if( ret ) {
		free(d);
        printf( "\nSKP_Silk_InitDecoder returned %d", ret );
		return NULL;
    }

	return d;
}

void free_decoder(void *d)
{
	if (d) free(d);
}

int decode_frame(void *d, int sample_rate, void *payload, int nbytes, void *out, int len)
{
	SKP_int16 ret = 0, used = 0;
	//SKP_int16 out[ ( ( FRAME_LENGTH_MS * MAX_API_FS_KHZ ) << 1 ) * MAX_INPUT_FRAMES ];

    SKP_SILK_SDK_DecControlStruct dc;
	dc.API_sampleRate = sample_rate;
    dc.framesPerPacket = 1;

    ret = SKP_Silk_SDK_Decode(d, &dc, 0, payload, nbytes, out, &used);
    if( ret ) {
       printf( "\nSKP_Silk_SDK_Decode returned %d", ret );
		return -1;
    }

	return used * 2;
}

