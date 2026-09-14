<?php

declare(strict_types=1);

namespace OCA\PcsSso\Listener;

use OCP\AppFramework\Http\ContentSecurityPolicy;
use OCP\EventDispatcher\Event;
use OCP\EventDispatcher\IEventListener;
use OCP\IConfig;
use OCP\Security\CSP\AddContentSecurityPolicyEvent;

/** @template-implements IEventListener<AddContentSecurityPolicyEvent> */
class CSPListener implements IEventListener {
	public function __construct(private IConfig $config) {
	}

	public function handle(Event $event): void {
		if (!($event instanceof AddContentSecurityPolicyEvent)) {
			return;
		}
		$origin = rtrim(trim($this->config->getSystemValueString('pcs_sso_allowed_origin', '')), '/');
		if (!preg_match('#^https://[a-z0-9.-]+(?::[0-9]+)?$#i', $origin)) {
			return;
		}
		$policy = new ContentSecurityPolicy();
		$policy->addAllowedFrameAncestorDomain($origin);
		$event->addPolicy($policy);
	}
}
