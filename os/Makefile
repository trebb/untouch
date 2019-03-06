BUILDROOT_TARBALL=buildroot-2018.11.3.tar.gz
BUILDROOT_DIR=buildroot-2018.11.3
BUILDROOT_EXTERNAL_CONFIG=br-output/.config
UI_BINARY=br-external/board/raspberrypi/rootfs_overlay/root/ui
SDCARD_IMG=br-output/images/sdcard.img

all: $(SDCARD_IMG)

$(BUILDROOT_TARBALL):
	wget https://buildroot.org/downloads/buildroot-2018.11.3.tar.gz

$(BUILDROOT_DIR): $(BUILDROOT_TARBALL)
	tar -xzf $<

$(BUILDROOT_EXTERNAL_CONFIG): $(BUILDROOT_DIR)
	$(MAKE) -C $(BUILDROOT_DIR) defconfig O=../br-output BR2_DEFCONFIG=../br-external/configs/novus-raspberrypi_defconfig

$(UI_BINARY): $(BUILDROOT_DIR)
	$(MAKE) -C ../software
	mkdir -p $(dir $(UI_BINARY))
	cp ../software/ui $(UI_BINARY)

$(SDCARD_IMG): $(BUILDROOT_EXTERNAL_CONFIG) $(UI_BINARY)
	$(MAKE) -C $(BUILDROOT_DIR) O=../br-output

clean:
	rm -rf $(BUILDROOT_DIR) br-output $(UI_BINARY)
