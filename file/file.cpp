#include <stdio.h>
#include <stdlib.h>

int splice(const char *filepath, long start, long end) {
    if (start >= end) {
        fprintf(stderr, "Invalid range: start must be less than end.\n");
        return 0;
    }

    FILE *file = fopen(filepath, "rb");

    if (!file) {
        perror("Failed to open file for reading");
        return 0;
    }

    fseek(file, 0, SEEK_END);

    long file_size = ftell(file);

    if (start < 0 || end > file_size) {
        fprintf(stderr, "Index range out of bounds.\n");
        fclose(file);
        return 0;
    }

    long slice_size = end - start;
    char *buffer = (char *)malloc(slice_size);

    if (!buffer) {
        fprintf(stderr, "Failed to allocate memory.\n");
        fclose(file);
        return 0;
    }

    fseek(file, start, SEEK_SET);
    fread(buffer, 1, slice_size, file);
    fclose(file);

    file = fopen(filepath, "wb");

    if (!file) {
        perror("Failed to open file for writing");
        free(buffer);
        return 0;
    }

    fwrite(buffer, 1, slice_size, file);
    fclose(file);
    free(buffer);
    return 1;
}
