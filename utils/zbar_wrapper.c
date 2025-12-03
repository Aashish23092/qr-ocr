#include <zbar.h>
#include <stdlib.h>
#include <string.h>

#define ZBAR_Y800 0x30303859

int decode_qr_zbar(unsigned char *gray_data, int width, int height,
                   unsigned char *output, int *output_len) {

    zbar_image_scanner_t *scanner = zbar_image_scanner_create();
    if (!scanner)
        return -1;

    zbar_image_scanner_set_config(scanner, 0, ZBAR_CFG_ENABLE, 1);

    zbar_image_t *img = zbar_image_create();
    if (!img) {
        zbar_image_scanner_destroy(scanner);
        return -2;
    }

    zbar_image_set_format(img, ZBAR_Y800);
    zbar_image_set_size(img, width, height);
    zbar_image_set_data(img, gray_data, width * height, NULL);

    int n = zbar_scan_image(scanner, img);
    if (n < 1) {
        zbar_image_destroy(img);
        zbar_image_scanner_destroy(scanner);
        return -3;
    }

    const zbar_symbol_t *sym = zbar_image_first_symbol(img);
    if (!sym) {
        zbar_image_destroy(img);
        zbar_image_scanner_destroy(scanner);
        return -4;
    }

    const char *raw = zbar_symbol_get_data(sym);
    unsigned int len = zbar_symbol_get_data_length(sym);
    if (len > 10240) len = 10240;

    memcpy(output, raw, len);
    *output_len = len;

    zbar_image_destroy(img);
    zbar_image_scanner_destroy(scanner);

    return 0;
}
