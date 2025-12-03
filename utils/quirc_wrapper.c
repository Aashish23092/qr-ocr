#include <quirc.h>
#include <stdlib.h>
#include <string.h>

// C function visible to CGO
int decode_qr_quirc(unsigned char *gray_data, int width, int height,
                    unsigned char *output, int *output_len) {

    struct quirc *qr = quirc_new();
    if (!qr) return -1;

    if (quirc_resize(qr, width, height) < 0) {
        quirc_destroy(qr);
        return -2;
    }

    int w, h;
    uint8_t *buf = quirc_begin(qr, &w, &h);
    if (!buf) {
        quirc_destroy(qr);
        return -3;
    }

    memcpy(buf, gray_data, width * height);
    quirc_end(qr);

    if (quirc_count(qr) < 1) {
        quirc_destroy(qr);
        return -4;
    }

    struct quirc_code code;
    struct quirc_data data;

    quirc_extract(qr, 0, &code);

    if (quirc_decode(&code, &data) != QUIRC_SUCCESS) {
        quirc_destroy(qr);
        return -5;
    }

    memcpy(output, data.payload, data.payload_len);
    *output_len = data.payload_len;

    quirc_destroy(qr);
    return 0;
}
