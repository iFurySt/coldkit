//go:build darwin && cgo

package keychain

/*
#cgo CFLAGS: -x objective-c -fblocks
#cgo LDFLAGS: -framework CoreFoundation -framework Foundation -framework Security -framework LocalAuthentication
#include <CoreFoundation/CoreFoundation.h>
#include <dispatch/dispatch.h>
#include <Foundation/Foundation.h>
#include <LocalAuthentication/LocalAuthentication.h>
#include <Security/Security.h>
#include <stdlib.h>

static void setString(CFMutableDictionaryRef dict, const void *key, const char *value) {
	CFStringRef str = CFStringCreateWithCString(kCFAllocatorDefault, value, kCFStringEncodingUTF8);
	if (str == NULL) {
		return;
	}
	CFDictionarySetValue(dict, key, str);
	CFRelease(str);
}

static CFMutableDictionaryRef baseQuery(const char *service, const char *account) {
	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		kCFAllocatorDefault,
		0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks
	);
	if (query == NULL) {
		return NULL;
	}
	CFDictionarySetValue(query, kSecClass, kSecClassGenericPassword);
	setString(query, kSecAttrService, service);
	setString(query, kSecAttrAccount, account);
	return query;
}

static int coldkitAuthorize(const char *prompt, char **errorMessage) {
	LAContext *context = [[LAContext alloc] init];
	NSString *reason = [NSString stringWithUTF8String:prompt];
	dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);
	__block int authorized = 0;
	__block NSString *failure = nil;

	[context evaluatePolicy:LAPolicyDeviceOwnerAuthentication localizedReason:reason reply:^(BOOL success, NSError *error) {
		if (success) {
			authorized = 1;
		} else if (error != nil) {
			failure = [[error localizedDescription] retain];
		}
		dispatch_semaphore_signal(semaphore);
	}];
	dispatch_semaphore_wait(semaphore, DISPATCH_TIME_FOREVER);

	if (!authorized && failure != nil && errorMessage != NULL) {
		const char *utf8 = [failure UTF8String];
		if (utf8 != NULL) {
			*errorMessage = strdup(utf8);
		}
		[failure release];
	}
	[context release];
	return authorized;
}

static OSStatus coldkitDeleteGenericPassword(const char *service, const char *account) {
	CFMutableDictionaryRef query = baseQuery(service, account);
	if (query == NULL) {
		return errSecParam;
	}
	OSStatus status = SecItemDelete(query);
	CFRelease(query);
	return status;
}

static OSStatus coldkitAddGenericPassword(
	const char *service,
	const char *account,
	const char *label,
	const char *comment,
	const UInt8 *secret,
	CFIndex secretLen
) {
	CFMutableDictionaryRef query = baseQuery(service, account);
	if (query == NULL) {
		return errSecParam;
	}
	setString(query, kSecAttrLabel, label);
	setString(query, kSecAttrDescription, "coldkit TRON private key");
	setString(query, kSecAttrComment, comment);

	CFDataRef data = CFDataCreate(kCFAllocatorDefault, secret, secretLen);
	if (data == NULL) {
		CFRelease(query);
		return errSecParam;
	}
	CFDictionarySetValue(query, kSecValueData, data);
	CFRelease(data);

	OSStatus status = SecItemAdd(query, NULL);
	CFRelease(query);
	return status;
}

static OSStatus coldkitCopyGenericPassword(
	const char *service,
	const char *account,
	const char *prompt,
	CFDataRef *result
) {
	CFMutableDictionaryRef query = baseQuery(service, account);
	if (query == NULL) {
		return errSecParam;
	}
	CFDictionarySetValue(query, kSecReturnData, kCFBooleanTrue);
	CFDictionarySetValue(query, kSecMatchLimit, kSecMatchLimitOne);

	CFTypeRef item = NULL;
	OSStatus status = SecItemCopyMatching(query, &item);
	CFRelease(query);
	if (status != errSecSuccess) {
		return status;
	}
	*result = (CFDataRef)item;
	return status;
}

static char *coldkitOSStatusMessage(OSStatus status) {
	CFStringRef message = SecCopyErrorMessageString(status, NULL);
	if (message == NULL) {
		return NULL;
	}
	CFIndex length = CFStringGetLength(message);
	CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;
	char *buffer = (char *)calloc(maxSize, sizeof(char));
	if (buffer == NULL) {
		CFRelease(message);
		return NULL;
	}
	if (!CFStringGetCString(message, buffer, maxSize, kCFStringEncodingUTF8)) {
		free(buffer);
		buffer = NULL;
	}
	CFRelease(message);
	return buffer;
}

static char *coldkitCFErrorMessage(CFErrorRef error) {
	if (error == NULL) {
		return NULL;
	}
	CFStringRef message = CFErrorCopyDescription(error);
	if (message == NULL) {
		return NULL;
	}
	CFIndex length = CFStringGetLength(message);
	CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;
	char *buffer = (char *)calloc(maxSize, sizeof(char));
	if (buffer == NULL) {
		CFRelease(message);
		return NULL;
	}
	if (!CFStringGetCString(message, buffer, maxSize, kCFStringEncodingUTF8)) {
		free(buffer);
		buffer = NULL;
	}
	CFRelease(message);
	return buffer;
}

static CFIndex coldkitDataLength(CFDataRef data) {
	return CFDataGetLength(data);
}

static const UInt8 *coldkitDataBytes(CFDataRef data) {
	return CFDataGetBytePtr(data);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

func storeSecret(account string, secret string, comment string) error {
	serviceC := C.CString(tronPrivateKeyService)
	accountC := C.CString(account)
	labelC := C.CString("coldkit TRON private key " + account)
	commentC := C.CString(comment)
	defer C.free(unsafe.Pointer(serviceC))
	defer C.free(unsafe.Pointer(accountC))
	defer C.free(unsafe.Pointer(labelC))
	defer C.free(unsafe.Pointer(commentC))

	deleteStatus := C.coldkitDeleteGenericPassword(serviceC, accountC)
	if deleteStatus != C.errSecSuccess && deleteStatus != C.errSecItemNotFound {
		return fmt.Errorf("replace existing macOS Keychain item: %s", osStatusError(deleteStatus))
	}

	secretBytes := []byte(secret)
	var secretPtr *C.UInt8
	if len(secretBytes) > 0 {
		secretPtr = (*C.UInt8)(unsafe.Pointer(&secretBytes[0]))
	}

	status := C.coldkitAddGenericPassword(
		serviceC,
		accountC,
		labelC,
		commentC,
		secretPtr,
		C.CFIndex(len(secretBytes)),
	)
	clearBytes(secretBytes)
	if status != C.errSecSuccess {
		return fmt.Errorf("store key in macOS Keychain: %s", osStatusError(status))
	}
	return nil
}

func loadSecret(account string) (string, error) {
	serviceC := C.CString(tronPrivateKeyService)
	accountC := C.CString(account)
	promptC := C.CString("Authorize coldkit to sign with TRON key " + account)
	defer C.free(unsafe.Pointer(serviceC))
	defer C.free(unsafe.Pointer(accountC))
	defer C.free(unsafe.Pointer(promptC))

	var authError *C.char
	if C.coldkitAuthorize(promptC, &authError) == 0 {
		if authError != nil {
			defer C.free(unsafe.Pointer(authError))
			return "", fmt.Errorf("authorize macOS Keychain access: %s", C.GoString(authError))
		}
		return "", errors.New("authorize macOS Keychain access: authentication failed")
	}

	var data C.CFDataRef
	status := C.coldkitCopyGenericPassword(serviceC, accountC, promptC, &data)
	if status != C.errSecSuccess {
		return "", fmt.Errorf("load key %q from macOS Keychain: %s", account, osStatusError(status))
	}
	defer C.CFRelease(C.CFTypeRef(data))

	length := int(C.coldkitDataLength(data))
	if length == 0 {
		return "", errors.New("macOS Keychain returned an empty secret")
	}
	bytes := C.GoBytes(unsafe.Pointer(C.coldkitDataBytes(data)), C.int(length))
	defer clearBytes(bytes)
	return string(bytes), nil
}

func deleteSecret(account string) error {
	serviceC := C.CString(tronPrivateKeyService)
	accountC := C.CString(account)
	defer C.free(unsafe.Pointer(serviceC))
	defer C.free(unsafe.Pointer(accountC))

	status := C.coldkitDeleteGenericPassword(serviceC, accountC)
	if status != C.errSecSuccess {
		return fmt.Errorf("delete key %q from macOS Keychain: %s", account, osStatusError(status))
	}
	return nil
}

func osStatusError(status C.OSStatus) string {
	message := C.coldkitOSStatusMessage(status)
	if message == nil {
		return fmt.Sprintf("OSStatus %d", int(status))
	}
	defer C.free(unsafe.Pointer(message))
	return C.GoString(message)
}

func clearBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}
