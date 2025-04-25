#include <stdio.h>
#include <stdlib.h>
#include <string.h>
// #include <iostream>


// using namespace std;



int splice(const char *filepath, double start, double end) {
    if (start >= end) {
        fprintf(stderr, "Invalid range: start must be less than end.\n");
        return 0;
    }

    FILE *file = fopen(filepath, "rb");

    if (!file) {
        perror("Failed to open file for reading");
        fprintf(stderr, filepath, "\n");
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

int splicePrefix(const char *filepath, const char *prefix, double backOffset) {
    FILE *f = fopen(filepath, "rb");
    if (!f) {
        perror("Opening file for read");
        return 0;
    }

    // determine file size
    if (fseek(f, 0, SEEK_END) != 0) {
        perror("Seeking to end");
        fclose(f);
        return 0;
    }
    long file_size = ftell(f);
    if (file_size < 0) {
        perror("Getting file size");
        fclose(f);
        return 0;
    }
    rewind(f);

    // read entire file into memory
    char *buf = (char *)malloc(file_size);
    if (!buf) {
        fprintf(stderr, "Out of memory\n");
        fclose(f);
        return 0;
    }
    if (fread(buf, 1, file_size, f) != (size_t)file_size) {
        perror("Reading file");
        free(buf);
        fclose(f);
        return 0;
    }
    fclose(f);

    // find the first occurrence of prefix in buffer
    size_t prefix_len = strlen(prefix);
    long offset = -1;
    for (long i = 0; i + (long)prefix_len <= file_size; i++) {
        if (memcmp(buf + i, prefix, prefix_len) == 0) {
            offset = i - backOffset;
            break;
        }
    }
    if (offset < 0) {
        fprintf(stderr, "Prefix not found in file\n");
        free(buf);
        return 0;
    }

    // open for truncation + write only up to the prefix (excluding prefix)
    f = fopen(filepath, "wb");
    if (!f) {
        perror("Opening file for write");
        free(buf);
        return 0;
    }
    size_t new_size = offset;
    if (fwrite(buf, 1, new_size, f) != new_size) {
        perror("Writing truncated data");
        free(buf);
        fclose(f);
        return 0;
    }

    free(buf);
    fclose(f);
    return 1;
}

int join(const char *dst_path, const char *src_path) {
    FILE *dst = fopen(dst_path, "ab");
    if (!dst) return 0;

    FILE *src = fopen(src_path, "rb");
    if (!src) {
        fclose(dst);
        return 0;
    }

    char buffer[4096];
    size_t bytes;
    while ((bytes = fread(buffer, 1, sizeof(buffer), src)) > 0) {
        if (fwrite(buffer, 1, bytes, dst) != bytes) {
            fclose(src);
            fclose(dst);
            return 0;
        }
    }

    fclose(src);
    fclose(dst);
    return 1;
}


int create(const char *filepath) {
    FILE *f = fopen(filepath, "wb");
    
    return !f ? 0 : 1;
}
