// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { ModuleWithProviders, NgModule } from '@angular/core';
import {
    ChildTranslateServiceConfig,
    provideChildTranslateService,
    provideTranslateService,
    RootTranslateServiceConfig,
    TranslateDirective,
    TranslatePipe,
} from '@ngx-translate/core';

type LegacyTranslateModuleConfig = RootTranslateServiceConfig & {
    defaultLanguage?: string;
    extend?: boolean;
};

@NgModule({
    imports: [TranslateDirective, TranslatePipe],
    exports: [TranslateDirective, TranslatePipe],
})
export class TranslateModule {
    static forRoot(
        config: LegacyTranslateModuleConfig = {}
    ): ModuleWithProviders<TranslateModule> {
        return {
            ngModule: TranslateModule,
            providers: provideTranslateService({
                ...config,
                fallbackLang: config.fallbackLang ?? config.defaultLanguage,
            }),
        };
    }

    static forChild(
        config: LegacyTranslateModuleConfig = {}
    ): ModuleWithProviders<TranslateModule> {
        const childConfig: ChildTranslateServiceConfig = {
            loader: config.loader,
            compiler: config.compiler,
            parser: config.parser,
            missingTranslationHandler: config.missingTranslationHandler,
        };

        return {
            ngModule: TranslateModule,
            providers: provideChildTranslateService(childConfig),
        };
    }
}
