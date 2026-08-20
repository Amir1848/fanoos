import { Directive, TemplateRef, ViewContainerRef, effect, inject, input } from '@angular/core';
import { AuthService } from './auth.service';

/**
 * Structural directive that renders its host content only when the logged-in
 * user has the given resource key (fetched once per session via
 * /api/auth/resources — see AuthService.ensureResourcesLoaded).
 *
 * Usage: <button *appAuthorize="'roles.create'">…</button>
 */
@Directive({ selector: '[appAuthorize]' })
export class AuthorizeDirective {
  private readonly templateRef = inject<TemplateRef<unknown>>(TemplateRef);
  private readonly viewContainer = inject(ViewContainerRef);
  private readonly auth = inject(AuthService);

  /** Resource key required to render the content, e.g. 'roles.create'. */
  readonly appAuthorize = input.required<string>();

  constructor() {
    effect(() => {
      // Empty key = no restriction (always visible).
      if (!this.appAuthorize() || this.auth.hasResource(this.appAuthorize())) {
        if (this.viewContainer.length === 0) {
          this.viewContainer.createEmbeddedView(this.templateRef);
        }
      } else {
        this.viewContainer.clear();
      }
    });
  }
}
