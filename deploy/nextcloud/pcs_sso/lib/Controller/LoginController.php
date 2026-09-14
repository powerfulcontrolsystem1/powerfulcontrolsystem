<?php

declare(strict_types=1);

namespace OCA\PcsSso\Controller;

use OCP\AppFramework\Controller;
use OCP\AppFramework\Http\RedirectResponse;
use OCP\IConfig;
use OCP\IRequest;
use OCP\ISession;
use OCP\IURLGenerator;
use OCP\IUserManager;
use OCP\IUserSession;

class LoginController extends Controller {
	private const AUDIENCE = 'pcs-nextcloud';
	private const MAX_LIFETIME_SECONDS = 60;

	public function __construct(
		string $appName,
		IRequest $request,
		private IConfig $config,
		private ISession $session,
		private IUserManager $userManager,
		private IUserSession $userSession,
		private IURLGenerator $urlGenerator,
	) {
		parent::__construct($appName, $request);
	}

	/**
	 * @PublicPage
	 * @NoCSRFRequired
	 * @UseSession
	 */
	public function start(string $token = ''): RedirectResponse {
		$claims = $this->validateToken($token);
		if ($claims === null) {
			return $this->loginPage();
		}

		$uid = (string)$claims['uid'];
		$user = $this->userManager->get($uid);
		if ($user === null || !$user->isEnabled()) {
			return $this->loginPage();
		}

		$this->session->regenerateId();
		$this->userSession->setUser($user);
		$this->session->set('loginname', $uid);
		$user->updateLastLoginTimestamp();

		return new RedirectResponse($this->urlGenerator->linkToRouteAbsolute('files.view.index'));
	}

	private function loginPage(): RedirectResponse {
		return new RedirectResponse($this->urlGenerator->linkToRouteAbsolute('core.login.showLoginForm'));
	}

	/** @return array<string,mixed>|null */
	private function validateToken(string $token): ?array {
		$secret = trim($this->config->getSystemValueString('pcs_sso_secret', ''));
		if (strlen($secret) < 32 || strlen($token) > 4096) {
			return null;
		}
		$parts = explode('.', trim($token));
		if (count($parts) !== 2 || $parts[0] === '' || $parts[1] === '') {
			return null;
		}
		$expected = $this->base64UrlEncode(hash_hmac('sha256', $parts[0], $secret, true));
		if (!hash_equals($expected, $parts[1])) {
			return null;
		}
		$payload = $this->base64UrlDecode($parts[0]);
		if ($payload === null) {
			return null;
		}
		try {
			$claims = json_decode($payload, true, 16, JSON_THROW_ON_ERROR);
		} catch (\JsonException) {
			return null;
		}
		if (!is_array($claims)) {
			return null;
		}

		$now = time();
		$uid = isset($claims['uid']) && is_string($claims['uid']) ? $claims['uid'] : '';
		$empresaId = isset($claims['empresa_id']) && is_int($claims['empresa_id']) ? $claims['empresa_id'] : 0;
		$issuedAt = isset($claims['iat']) && is_int($claims['iat']) ? $claims['iat'] : 0;
		$expiresAt = isset($claims['exp']) && is_int($claims['exp']) ? $claims['exp'] : 0;
		$nonce = isset($claims['nonce']) && is_string($claims['nonce']) ? $claims['nonce'] : '';
		if (($claims['v'] ?? null) !== 1 || ($claims['aud'] ?? null) !== self::AUDIENCE) {
			return null;
		}
		if ($empresaId <= 0 || $uid !== 'pcs_empresa_' . $empresaId || !preg_match('/^pcs_empresa_[1-9][0-9]*$/', $uid)) {
			return null;
		}
		if ($issuedAt > $now + 5 || $expiresAt < $now || $expiresAt <= $issuedAt || $expiresAt - $issuedAt > self::MAX_LIFETIME_SECONDS) {
			return null;
		}
		if (strlen($nonce) < 16 || strlen($nonce) > 64) {
			return null;
		}
		return $claims;
	}

	private function base64UrlEncode(string $value): string {
		return rtrim(strtr(base64_encode($value), '+/', '-_'), '=');
	}

	private function base64UrlDecode(string $value): ?string {
		$padding = strlen($value) % 4;
		if ($padding !== 0) {
			$value .= str_repeat('=', 4 - $padding);
		}
		$decoded = base64_decode(strtr($value, '-_', '+/'), true);
		return $decoded === false ? null : $decoded;
	}
}
