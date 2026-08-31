/* ============================================================
   Huz CCTV — Client-side internationalization (i18n)
   - English is the default language; Vietnamese is also available.
   - Detects the browser language on first visit, persists the
     choice in localStorage ("huz_lang").
   - Applies [data-i18n] / [data-i18n-placeholder] / [data-i18n-aria]
     attributes and updates <title> / <html lang>.
   - Emits a "huz:langchange" event so pages can re-render
     dynamic text.
   ============================================================ */
(function () {
  'use strict';

  var STORAGE_KEY = 'huz_lang';
  var SUPPORTED = ['en', 'vi'];

  var DICTS = {
    /* ------------------------------------------------ English (default) */
    en: {
      // Header / navigation
      'nav.dashboard': 'Dashboard',
      'nav.network': 'Network Devices',
      'nav.camera': 'Camera',
      'nav.phone': 'Phone Remote',
      'header.brand.sub': 'Smart Surveillance',
      'header.menu': 'Menu',
      'header.logout': 'Log out',
      'header.loggingOut': 'Signing out…',
      'header.login': 'Sign in',
      'header.language': 'Language',
      'header.donate.text': 'Support developer for future update and features',

      // Footer
      'footer.version': 'Server version',

      // API / backend error codes
      'error.invalid_input': 'Please enter username and password.',
      'error.rate_limited': 'Too many failed login attempts. Please try again in 15 minutes.',
      'error.invalid_credentials': 'Invalid username or password.',
      'error.session_error': 'Could not create a login session.',
      'error.unauthorized': 'Your session has expired. Please sign in again.',
      'error.request_failed': 'Request failed.',
      'error.no_network_iface': 'No active network interface found.',
      'error.scan_failed': 'Could not get the network device list.',
      'error.missing_password': 'Missing password information.',
      'error.password_too_short': 'New password must be at least 8 characters.',
      'error.wrong_current_password': 'Current password is incorrect.',

      // Login page
      'login.title': 'Sign in · Huz CCTV',
      'login.aria': 'Sign in',
      'login.subtitle': 'Sign in to continue managing your surveillance system',
      'login.username': 'Username',
      'login.username.placeholder': 'Enter your username',
      'login.password': 'Password',
      'login.password.placeholder': 'Enter your password',
      'login.toggle.aria': 'Show / hide password',
      'login.toggle.show': 'Show password',
      'login.toggle.hide': 'Hide password',
      'login.remember': 'Remember me on this browser',
      'login.remember.hint': 'Your session will be kept for a long time, so you will not need to sign in again every time you open the page.',
      'login.submit': 'Sign in',
      'login.signingIn': 'Signing in…',
      'login.divider': 'Surveillance system',
      'login.footer': '© {year} Huz CCTV · Safe · Secure · Reliable',
      'login.error.generic': 'Something went wrong, please try again.',
      'login.error.required': 'Please enter both username and password.',
      'login.error.failed': 'Sign in failed.',
      // Dashboard
      'dash.title': 'Dashboard · Huz CCTV',
      'dash.pageTitle': 'Dashboard',
      'dash.sub': 'Welcome to the Huz CCTV surveillance system.',
      'dash.greeting.morning': 'Good morning',
      'dash.greeting.afternoon': 'Good afternoon',
      'dash.greeting.evening': 'Good evening',
      'dash.greeting.admin': 'administrator',
      'dash.greeting.format': '{greeting}, {username}! The Huz CCTV surveillance system is ready.',
      'dash.networkDevices': 'Network Devices',
      'dash.viewCamera': 'View cameras',
      'dash.phoneRemote': 'Phone Remote',
      'dash.stat.server': 'Server',
      'dash.stat.server.checking': 'Checking…',
      'dash.stat.server.foot': 'Checking connection to the backend',
      'dash.stat.server.ok': 'Running',
      'dash.stat.server.okFoot': 'Backend is responding normally',
      'dash.stat.server.err': 'Connection lost',
      'dash.stat.server.errFoot': 'Cannot reach the backend',
      'dash.stat.devices': 'Network devices',
      'dash.stat.devices.foot': 'Devices found on the local network',
      'dash.stat.devices.updated': 'Updated {time}',
      'dash.stat.devices.error': 'Could not scan the local network',
      'dash.stat.user': 'User',
      'dash.stat.user.foot': 'Active signed-in session',
      'dash.server.ip': 'IP Address',
      'dash.server.port': 'Port',
      'dash.server.hostname': 'Hostname',
      'dash.server.uptime': 'Uptime',
      'dash.server.os': 'Operating system',
      'dash.server.version': 'App version',
      'dash.server.go': 'Go · CPU',
      'dash.server.started': 'Started at',
      'dash.quickstart': 'Quick start',
      'dash.quickstart.desc': 'Huz CCTV connects directly to cameras on your local network over WebRTC. Video streams travel directly between the device and your browser — no intermediaries.',
      'dash.viewCameras': 'View cameras',
      'dash.scanDevices': 'Scan network devices',
      'dash.sysinfo': 'System information',
      'dash.sysinfo.backend': 'Backend',
      'dash.sysinfo.security': 'Security',
      'dash.sysinfo.ui': 'UI version',
      'dash.uptime.day': '{n}d',
      'dash.uptime.hour': '{n}h',
      'dash.uptime.min': '{n}m',
      'dash.uptime.sec': '{n}s',


      // Devices page
      'devices.title': 'Network Devices · Huz CCTV',
      'devices.pageTitle': 'Network Devices',
      'devices.sub': 'Devices discovered on the current local network.',
      'devices.count': ['{n} device', '{n} devices'],
      'devices.scan': 'Scan again',
      'devices.list': 'Device list',
      'devices.updatedAt': '· Updated at {time}',
      'devices.loading': 'Scanning the local network…',
      'devices.loading.sub': 'This may take a few seconds',
      'devices.empty': 'No devices found',
      'devices.empty.sub': 'Check your network interface and try scanning again.',
      'devices.error': 'Unable to scan devices',
      'devices.error.sub': 'An error occurred.',
      'devices.tryAgain': 'Try again',
      'devices.th.ip': 'IP Address',
      'devices.th.mac': 'MAC Address',
      'devices.th.vendor': 'Vendor',
      'devices.th.hostname': 'Hostname',
      'devices.th.iface': 'Interface',
      'devices.th.state': 'State',
      'devices.unknown': 'Unknown',
      'devices.online': 'Online',
      'devices.offline': 'Offline',

      // Camera page
      'camera.title': 'Camera · Huz CCTV',
      'camera.awaiting': 'Awaiting',
      'camera.refresh': 'Refresh',
      'camera.connectAll': 'Connect all',
      'camera.stat.total': 'Total devices',
      'camera.stat.streaming': 'Streaming',
      'camera.stat.connection': 'Connection state',
      'camera.online': 'Online',
      'camera.offline': 'Offline',
      'camera.empty.title': 'No camera devices connected yet.',
      'camera.empty.sub': 'When devices connect, video streams will appear here automatically.',
      'camera.status.connected': 'Connected',
      'camera.status.ready': 'Ready',
      'camera.status.disconnected': 'Disconnected',
      'camera.status.error': 'Connection error',
      'camera.status.wsError': 'WebSocket error',
      'camera.status.refreshing': 'Refreshing list',
      'camera.status.reconnectingAll': 'Reconnecting all…',
      'camera.badge.waiting': 'Waiting…',
      'camera.badge.awaitingReply': 'Waiting for reply…',
      'camera.badge.connected': 'Connected',
      'camera.badge.live': 'Live',
      'camera.badge.disconnected': 'Disconnected',
      'camera.overlay.connecting': 'Connecting…',
      'camera.overlay.connectFailed': 'Unable to connect to the video stream',
      'camera.overlay.streamLost': 'Video stream lost',
      'camera.reconnect': 'Reconnect',
      'camera.viewerName': 'CCTV Browser',
      'camera.deviceName': 'Camera',
      'camera.zoom.title': 'Digital zoom',
      'camera.zoom.in': 'Zoom in',
      'camera.zoom.out': 'Zoom out',
      'camera.zoom.reset': 'Reset zoom',
      'camera.ptz.title': 'Camera control',
      'camera.ptz.up': 'Tilt up',
      'camera.ptz.down': 'Tilt down',
      'camera.ptz.left': 'Pan left',
      'camera.ptz.right': 'Pan right',
      'camera.ptz.home': 'Return to home position',
      'camera.dev.zoom': 'Zoom',
      'camera.dev.reset': 'Reset',
      'camera.dev.torch': 'Torch',
      'camera.dev.switch': 'Camera',
      'camera.dev.snapshot': 'Photo',
      'camera.dev.motion': 'Motion',
      'camera.dev.focus': 'Focus',
      'camera.motion': 'Motion',
      'camera.snapshot.close': 'Close snapshot',
      // Phone remote page
      'phone.title': 'Phone Remote · Huz CCTV',
      'phone.pageTitle': 'Phone Remote',
      'phone.sub': 'Manage and control your connected Android phones.',
      'phone.viewerName': 'Phone Remote',
      'phone.select.placeholder': 'Select a device…',
      'phone.refresh': 'Refresh',
      'phone.empty.title': 'No Android device connected',
      'phone.empty.sub': 'When a device connects it appears here. Select it to view and control its screen.',
      'phone.status.awaiting': 'Awaiting',
      'phone.status.ready': 'Ready',
      'phone.status.connecting': 'Connecting…',
      'phone.status.streaming': 'Streaming',
      'phone.status.noDevice': 'No device selected',
      'phone.status.disconnected': 'Disconnected',
      'phone.overlay.connecting': 'Connecting to the device…',
      'phone.overlay.awaitingShare': 'Connected — waiting for the screen share. Tap “Start screen share” and confirm on the phone.',
      'phone.overlay.connectFailed': 'Unable to connect to the device',
      'phone.overlay.streamLost': 'Video stream lost',
      'phone.section.screen': 'Screen sharing',
      'phone.section.nav': 'Navigation',
      'phone.section.power': 'Screen power',
      'phone.section.media': 'Media',
      'phone.screen.start': 'Start screen share',
      'phone.screen.stop': 'Stop screen share',
      'phone.screenOn': 'Turn screen on',
      'phone.screenOff': 'Turn screen off',
      'phone.key.back': 'Back',
      'phone.key.home': 'Home',
      'phone.key.recent': 'Recent',
      'phone.key.menu': 'Menu',
      'phone.key.volumeUp': 'Vol +',
      'phone.key.volumeDown': 'Vol −',
      'phone.snapshot': 'Snapshot',
      'phone.snapshot.close': 'Close image',
      'phone.screen.offNotice': 'Screen is off — live preview may pause',
      'phone.gesture.hint': 'Tap, drag or hold on the phone screen to control it',
      'phone.toast.noDevice': 'Select a device first',
      'ws.unauthorized': 'You need to sign in to view cameras.',
      'ws.replaced': 'The device reconnected in a new session; closing the old connection.',
      'ws.target_gone': 'The target device is no longer connected.'
    },
    /* ------------------------------------------------ Vietnamese */
    vi: {
      'nav.dashboard': 'Tổng quan',
      'nav.network': 'Thiết bị mạng',
      'nav.camera': 'Camera',
      'nav.phone': 'Điều khiển điện thoại',
      'header.brand.sub': 'Giám sát thông minh',
      'header.menu': 'Menu',
      'header.logout': 'Đăng xuất',
      'header.loggingOut': 'Đang xử lý…',
      'header.login': 'Đăng nhập',
      'header.language': 'Ngôn ngữ',
      'header.donate.text': 'Hỗ trợ nhà phát triển cho các bản cập nhật và tính năng trong tương lai',
      'footer.version': 'Phiên bản server',

      'error.invalid_input': 'Vui lòng nhập tên đăng nhập và mật khẩu.',
      'error.rate_limited': 'Quá nhiều lần đăng nhập sai, vui lòng thử lại sau 15 phút.',
      'error.invalid_credentials': 'Tên đăng nhập hoặc mật khẩu không đúng.',
      'error.session_error': 'Không thể tạo phiên đăng nhập.',
      'error.unauthorized': 'Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.',
      'error.request_failed': 'Yêu cầu thất bại.',
      'error.no_network_iface': 'Không tìm thấy card mạng nào đang hoạt động.',
      'error.scan_failed': 'Không thể lấy danh sách thiết bị mạng.',
      'error.missing_password': 'Thiếu thông tin mật khẩu.',
      'error.password_too_short': 'Mật khẩu mới phải có ít nhất 8 ký tự.',
      'error.wrong_current_password': 'Mật khẩu hiện tại không đúng.',

      'login.title': 'Đăng nhập · Huz CCTV',
      'login.aria': 'Đăng nhập',
      'login.subtitle': 'Đăng nhập để tiếp tục quản lý hệ thống giám sát',
      'login.username': 'Tên đăng nhập',
      'login.username.placeholder': 'Nhập tên đăng nhập',
      'login.password': 'Mật khẩu',
      'login.password.placeholder': 'Nhập mật khẩu',
      'login.toggle.aria': 'Hiện / ẩn mật khẩu',
      'login.toggle.show': 'Hiện mật khẩu',
      'login.toggle.hide': 'Ẩn mật khẩu',
      'login.remember': 'Ghi nhớ đăng nhập trên trình duyệt này',
      'login.remember.hint': 'Phiên đăng nhập sẽ được duy trì lâu dài, bạn không cần đăng nhập lại mỗi lần mở trang.',
      'login.submit': 'Đăng nhập',
      'login.signingIn': 'Đang đăng nhập…',
      'login.divider': 'Hệ thống giám sát',
      'login.footer': '© {year} Huz CCTV · An toàn · Bảo mật · Ổn định',
      'login.error.generic': 'Đã có lỗi xảy ra, vui lòng thử lại.',
      'login.error.required': 'Vui lòng nhập đầy đủ tên đăng nhập và mật khẩu.',
      'login.error.failed': 'Đăng nhập thất bại.',

      'dash.title': 'Tổng quan · Huz CCTV',
      'dash.pageTitle': 'Tổng quan',
      'dash.sub': 'Chào mừng bạn đến với hệ thống giám sát Huz CCTV.',
      'dash.greeting.morning': 'Chào buổi sáng',
      'dash.greeting.afternoon': 'Chào buổi chiều',
      'dash.greeting.evening': 'Chào buổi tối',
      'dash.greeting.admin': 'quản trị viên',
      'dash.greeting.format': '{greeting}, {username}! Hệ thống giám sát Huz CCTV đang sẵn sàng.',
      'dash.networkDevices': 'Thiết bị mạng',
      'dash.viewCamera': 'Xem camera',
      'dash.phoneRemote': 'Điều khiển điện thoại',
      'dash.stat.server': 'Máy chủ',
      'dash.stat.server.checking': 'Đang kiểm tra…',
      'dash.stat.server.foot': 'Kiểm tra kết nối tới backend',
      'dash.stat.server.ok': 'Hoạt động',
      'dash.stat.server.okFoot': 'Backend phản hồi bình thường',
      'dash.stat.server.err': 'Mất kết nối',
      'dash.stat.server.errFoot': 'Không thể liên hệ với backend',
      'dash.stat.devices': 'Thiết bị mạng',
      'dash.stat.devices.foot': 'Số thiết bị tìm thấy trên mạng LAN',
      'dash.stat.devices.updated': 'Cập nhật {time}',
      'dash.stat.devices.error': 'Chưa quét được mạng LAN',
      'dash.stat.user': 'Người dùng',
      'dash.stat.user.foot': 'Phiên đăng nhập đang hoạt động',
      'dash.server.ip': 'Địa chỉ IP',
      'dash.server.port': 'Cổng',
      'dash.server.hostname': 'Tên máy chủ',
      'dash.server.uptime': 'Thời gian hoạt động',
      'dash.server.os': 'Hệ điều hành',
      'dash.server.version': 'Phiên bản ứng dụng',
      'dash.server.go': 'Go · CPU',
      'dash.server.started': 'Khởi động lúc',
      'dash.quickstart': 'Bắt đầu nhanh',
      'dash.quickstart.desc': 'Huz CCTV kết nối trực tiếp với camera trên mạng LAN qua WebRTC. Luồng video được truyền trực tiếp giữa thiết bị và trình duyệt, không qua trung gian.',
      'dash.viewCameras': 'Xem camera',
      'dash.scanDevices': 'Quét thiết bị mạng',
      'dash.sysinfo': 'Thông tin hệ thống',
      'dash.sysinfo.backend': 'Backend',
      'dash.sysinfo.security': 'Bảo mật',
      'dash.sysinfo.ui': 'Phiên bản giao diện',
      'dash.uptime.day': '{n} ngày',
      'dash.uptime.hour': '{n} giờ',
      'dash.uptime.min': '{n} phút',
      'dash.uptime.sec': '{n} giây',

      'devices.title': 'Thiết bị mạng · Huz CCTV',
      'devices.pageTitle': 'Thiết bị mạng',
      'devices.sub': 'Danh sách thiết bị được phát hiện trên mạng LAN hiện tại.',
      'devices.count': '{n} thiết bị',
      'devices.scan': 'Quét lại',
      'devices.list': 'Danh sách thiết bị',
      'devices.updatedAt': '· Cập nhật lúc {time}',
      'devices.loading': 'Đang quét mạng LAN…',
      'devices.loading.sub': 'Quá trình này có thể mất vài giây',
      'devices.empty': 'Không tìm thấy thiết bị',
      'devices.empty.sub': 'Kiểm tra card mạng và thử quét lại.',
      'devices.error': 'Không thể quét thiết bị',
      'devices.error.sub': 'Đã có lỗi xảy ra.',
      'devices.tryAgain': 'Thử lại',
      'devices.th.ip': 'Địa chỉ IP',
      'devices.th.mac': 'Địa chỉ MAC',
      'devices.th.vendor': 'Nhà sản xuất',
      'devices.th.hostname': 'Hostname',
      'devices.th.iface': 'Giao diện',
      'devices.th.state': 'Trạng thái',
      'devices.unknown': 'Không rõ',
      'devices.online': 'Trực tuyến',
      'devices.offline': 'Ngoại tuyến',

      'camera.title': 'Camera · Huz CCTV',
      'camera.awaiting': 'Đang chờ',
      'camera.refresh': 'Làm mới',
      'camera.connectAll': 'Kết nối tất cả',
      'camera.stat.total': 'Tổng thiết bị',
      'camera.stat.streaming': 'Đang xem',
      'camera.stat.connection': 'Trạng thái kết nối',
      'camera.online': 'Trực tuyến',
      'camera.offline': 'Ngoại tuyến',
      'camera.empty.title': 'Chưa có thiết bị camera nào kết nối.',
      'camera.empty.sub': 'Khi thiết bị kết nối, luồng video sẽ tự động hiển thị tại đây.',
      'camera.status.connected': 'Đã kết nối',
      'camera.status.ready': 'Sẵn sàng',
      'camera.status.disconnected': 'Mất kết nối',
      'camera.status.error': 'Lỗi kết nối',
      'camera.status.wsError': 'Lỗi WebSocket',
      'camera.status.refreshing': 'Làm mới danh sách',
      'camera.status.reconnectingAll': 'Đang kết nối lại tất cả…',
      'camera.badge.waiting': 'Đang chờ…',
      'camera.badge.awaitingReply': 'Đang chờ phản hồi…',
      'camera.badge.connected': 'Đã kết nối',
      'camera.badge.live': 'Đang xem trực tiếp',
      'camera.badge.disconnected': 'Mất kết nối',
      'camera.overlay.connecting': 'Đang kết nối…',
      'camera.overlay.connectFailed': 'Không thể kết nối luồng video',
      'camera.overlay.streamLost': 'Mất kết nối luồng video',
      'camera.reconnect': 'Kết nối lại',
      'camera.viewerName': 'Trình duyệt CCTV',
      'camera.deviceName': 'Camera',
      'camera.zoom.title': 'Thu phóng',
      'camera.zoom.in': 'Phóng to',
      'camera.zoom.out': 'Thu nhỏ',
      'camera.zoom.reset': 'Đặt lại zoom',
      'camera.ptz.title': 'Điều khiển camera',
      'camera.ptz.up': 'Di chuyển lên',
      'camera.ptz.down': 'Di chuyển xuống',
      'camera.ptz.left': 'Di chuyển trái',
      'camera.ptz.right': 'Di chuyển phải',
      'camera.ptz.home': 'Về vị trí gốc',
      'camera.dev.zoom': 'Zoom',
      'camera.dev.reset': 'Đặt lại',
      'camera.dev.torch': 'Đèn pin',
      'camera.dev.switch': 'Đổi cam',
      'camera.dev.snapshot': 'Ảnh',
      'camera.dev.motion': 'Chuyển động',
      'camera.dev.focus': 'Lấy nét',
      'camera.motion': 'Chuyển động',
      'camera.snapshot.close': 'Đóng ảnh',
      // Trang điều khiển điện thoại
      'phone.title': 'Điều khiển điện thoại · Huz CCTV',
      'phone.pageTitle': 'Điều khiển điện thoại',
      'phone.sub': 'Quản lý và điều khiển các điện thoại Android đang kết nối.',
      'phone.viewerName': 'Điều khiển từ xa',
      'phone.select.placeholder': 'Chọn thiết bị…',
      'phone.refresh': 'Làm mới',
      'phone.empty.title': 'Chưa có điện thoại Android nào kết nối',
      'phone.empty.sub': 'Khi thiết bị kết nối sẽ hiện ở đây. Chọn thiết bị để xem và điều khiển màn hình.',
      'phone.status.awaiting': 'Đang chờ',
      'phone.status.ready': 'Sẵn sàng',
      'phone.status.connecting': 'Đang kết nối…',
      'phone.status.streaming': 'Đang phát',
      'phone.status.noDevice': 'Chưa chọn thiết bị',
      'phone.status.disconnected': 'Mất kết nối',
      'phone.overlay.connecting': 'Đang kết nối tới thiết bị…',
      'phone.overlay.awaitingShare': 'Đã kết nối — đang chờ chia sẻ màn hình. Nhấn “Bắt đầu chia sẻ màn hình” và xác nhận trên điện thoại.',
      'phone.overlay.connectFailed': 'Không kết nối được tới thiết bị',
      'phone.overlay.streamLost': 'Mất luồng video',
      'phone.section.screen': 'Chia sẻ màn hình',
      'phone.section.nav': 'Điều hướng',
      'phone.section.power': 'Nguồn màn hình',
      'phone.section.media': 'Phương tiện',
      'phone.screen.start': 'Bắt đầu chia sẻ màn hình',
      'phone.screen.stop': 'Dừng chia sẻ màn hình',
      'phone.screenOn': 'Bật màn hình',
      'phone.screenOff': 'Tắt màn hình',
      'phone.key.back': 'Quay lại',
      'phone.key.home': 'Trang chủ',
      'phone.key.recent': 'Gần đây',
      'phone.key.menu': 'Menu',
      'phone.key.volumeUp': 'Vol +',
      'phone.key.volumeDown': 'Vol −',
      'phone.snapshot': 'Ảnh',
      'phone.snapshot.close': 'Đóng ảnh',
      'phone.screen.offNotice': 'Màn hình đang tắt — bản xem trước có thể tạm dừng',
      'phone.gesture.hint': 'Chạm, kéo hoặc nhấn giữ trên màn hình điện thoại để điều khiển',
      'phone.toast.noDevice': 'Hãy chọn thiết bị trước',
      'ws.unauthorized': 'Bạn cần đăng nhập để xem camera.',
      'ws.replaced': 'Thiết bị đã kết nối lại ở phiên mới, đóng kết nối cũ.',
      'ws.target_gone': 'Thiết bị đích không còn kết nối.'
    }
  };

  function detectLang() {
    var saved = null;
    try {
      saved = localStorage.getItem(STORAGE_KEY);
    } catch (_e) {
      saved = null;
    }
    if (saved && SUPPORTED.indexOf(saved) !== -1) {
      return saved;
    }
    var nav = ((navigator.language || navigator.userLanguage) || 'en').toLowerCase();
    return nav.indexOf('vi') === 0 ? 'vi' : 'en';
  }

  var current = detectLang();

  /* Translate a key. Supports {placeholder} interpolation and plural
     arrays [singular, plural] selected by the {n} parameter. */
  function t(key, params) {
    var value = (DICTS[current] && DICTS[current][key]);
    if (value === undefined) value = DICTS.en[key];
    if (value === undefined) return key;

    if (params && params.n !== undefined && Array.isArray(value)) {
      value = value[Math.abs(Number(params.n)) === 1 ? 0 : 1];
    }
    if (Array.isArray(value)) value = value[0];

    if (params) {
      for (var k in params) {
        if (Object.prototype.hasOwnProperty.call(params, k)) {
          value = value.replace(new RegExp('\\{' + k + '\\}', 'g'), String(params[k]));
        }
      }
    }
    return value;
  }

  function lang() {
    return current;
  }

  function locale() {
    return current === 'vi' ? 'vi-VN' : 'en-US';
  }

  function applyTranslations(root) {
    var base = root || document;
    var i;

    var textNodes = base.querySelectorAll('[data-i18n]');
    for (i = 0; i < textNodes.length; i++) {
      var el = textNodes[i];
      var attr = el.getAttribute('data-i18n-attr');
      var key = el.getAttribute('data-i18n');
      if (attr && attr !== 'text') {
        el.setAttribute(attr, t(key));
      } else {
        el.textContent = t(key);
      }
    }

    var phNodes = base.querySelectorAll('[data-i18n-placeholder]');
    for (i = 0; i < phNodes.length; i++) {
      phNodes[i].setAttribute('placeholder', t(phNodes[i].getAttribute('data-i18n-placeholder')));
    }

    var ariaNodes = base.querySelectorAll('[data-i18n-aria]');
    for (i = 0; i < ariaNodes.length; i++) {
      ariaNodes[i].setAttribute('aria-label', t(ariaNodes[i].getAttribute('data-i18n-aria')));
    }

    var titleEl = base.querySelector('title');
    if (titleEl && titleEl.getAttribute('data-i18n')) {
      document.title = t(titleEl.getAttribute('data-i18n'));
    }
  }

  function setLang(next) {
    if (SUPPORTED.indexOf(next) === -1) return;
    current = next;
    try {
      localStorage.setItem(STORAGE_KEY, next);
    } catch (_e) {
      /* ignore */
    }
    document.documentElement.lang = next;
    applyTranslations(document);
    document.dispatchEvent(
      new CustomEvent('huz:langchange', { detail: { lang: next } })
    );
  }

  document.documentElement.lang = current;
  applyTranslations(document);

  window.I18N = {
    t: t,
    lang: lang,
    locale: locale,
    setLang: setLang,
    apply: applyTranslations
  };
})();

