BUILDROOT_VERSION=2018.11.3
BUILDROOT_TARBALL=buildroot-$(BUILDROOT_VERSION).tar.gz
BUILDROOT_DIR=buildroot-$(BUILDROOT_VERSION)
BUILDROOT_EXTERNAL_CONFIG=br-output/.config
UI_BINARY=br-external/board/raspberrypi/rootfs_overlay/root/ui
UI_BINARY_ORIGINAL=../software/ui
SDCARD_IMG=br-output/images/sdcard.img

all: $(SDCARD_IMG)

$(BUILDROOT_TARBALL):
	wget https://buildroot.org/downloads/buildroot-$(BUILDROOT_VERSION).tar.gz

$(BUILDROOT_DIR): $(BUILDROOT_TARBALL)
	tar -xzf $<

$(BUILDROOT_EXTERNAL_CONFIG): $(BUILDROOT_DIR)
	$(MAKE) -C $(BUILDROOT_DIR) defconfig O=../br-output BR2_DEFCONFIG=../br-external/configs/novus-raspberrypi_defconfig

$(UI_BINARY_ORIGINAL):
	$(MAKE) -C ../software

$(UI_BINARY): $(BUILDROOT_DIR) $(UI_BINARY_ORIGINAL)
	mkdir -p $(dir $(UI_BINARY))
	cp $(UI_BINARY_ORIGINAL) $(UI_BINARY)

$(SDCARD_IMG): $(BUILDROOT_EXTERNAL_CONFIG) $(UI_BINARY)
	$(MAKE) -C $(BUILDROOT_DIR) O=../br-output

clean:
	rm -rf $(BUILDROOT_DIR) br-output $(UI_BINARY)
