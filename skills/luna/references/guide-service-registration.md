## Service registration

`new ServiceManager(Log).AddDalamudServices(pi)`, `AddIServices(assembly)`, `BuildProvider()`,
then `GetService<ImSharpDalamudContext>()` and `EnsureRequiredServices(cancellationToken)`.
`AddSingleton<T>()` and `AddScoped<T>()` use the constructor with the **most parameters**. Every
`IService` type picks one lifetime: `ISingletonService`, `IScopedService` (which includes
`IUiService` and `IApiService`) or `IConstructedService`. → `references/topic-services.md`
