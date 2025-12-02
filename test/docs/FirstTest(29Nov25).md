# GoRoutinesManager Test Suite

## Overview

Comprehensive test suite for the GoRoutinesManager project covering all three manager implementations: GlobalManager, AppManager, and LocalManager.

**Total Tests: 30**
**Status: ✅ All Passing**

---

## Test Files

* GlobalManager_test.go - 10 tests
* AppManager_test.go - 9 tests
* LocalManager_test.go - 10 tests

---

## Running Tests

<pre><div node="[object Object]" class="relative whitespace-pre-wrap word-break-all p-3 my-2 rounded-sm bg-list-hover-subtle"><div><div class="code-block"><div class="code-line" data-line-number="1"><div class="line-content"><span class="mtk5"># Run all tests</span></div></div><div class="code-line" data-line-number="2"><div class="line-content"><span class="mtk16">go</span><span class="mtk1"></span><span class="mtk12">test</span><span class="mtk1"></span><span class="mtk6">-v</span><span class="mtk1"></span><span class="mtk12">./Tests/</span></div></div><div class="code-line" data-line-number="3"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="4"><div class="line-content"><span class="mtk5"># Run specific test file</span></div></div><div class="code-line" data-line-number="5"><div class="line-content"><span class="mtk16">go</span><span class="mtk1"></span><span class="mtk12">test</span><span class="mtk1"></span><span class="mtk6">-v</span><span class="mtk1"></span><span class="mtk12">./Tests/GlobalManager_test.go</span></div></div><div class="code-line" data-line-number="6"><div class="line-content"><span class="mtk16">go</span><span class="mtk1"></span><span class="mtk12">test</span><span class="mtk1"></span><span class="mtk6">-v</span><span class="mtk1"></span><span class="mtk12">./Tests/AppManager_test.go</span></div></div><div class="code-line" data-line-number="7"><div class="line-content"><span class="mtk16">go</span><span class="mtk1"></span><span class="mtk12">test</span><span class="mtk1"></span><span class="mtk6">-v</span><span class="mtk1"></span><span class="mtk12">./Tests/LocalManager_test.go</span></div></div><div class="code-line" data-line-number="8"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="9"><div class="line-content"><span class="mtk5"># Run specific test</span></div></div><div class="code-line" data-line-number="10"><div class="line-content"><span class="mtk16">go</span><span class="mtk1"></span><span class="mtk12">test</span><span class="mtk1"></span><span class="mtk6">-v</span><span class="mtk1"></span><span class="mtk12">./Tests/</span><span class="mtk1"></span><span class="mtk6">-run</span><span class="mtk1"></span><span class="mtk12">TestGlobalManager_Init</span></div></div></div></div></div></pre>

---

## Test Coverage

### GlobalManager Tests (10)

| Test                                                   | Description                                              |
| ------------------------------------------------------ | -------------------------------------------------------- |
| **TestGlobalManager_Init**                       | Tests initialization and idempotency                     |
| **TestGlobalManager_GetAllAppManagers_Empty**    | Verifies empty state handling                            |
| **TestGlobalManager_GetAppManagerCount**         | Tests app manager counting                               |
| **TestGlobalManager_GetAllAppManagers_WithApps** | Tests with multiple apps                                 |
| **TestGlobalManager_GetAllLocalManagers**        | Tests local manager aggregation                          |
| **TestGlobalManager_GetLocalManagerCount**       | Tests local manager counting                             |
| **TestGlobalManager_GetAllGoroutines**           | Tests goroutine aggregation                              |
| **TestGlobalManager_GetGoroutineCount**          | Tests goroutine counting                                 |
| **TestGlobalManager_MultipleAppsAndLocals**      | Tests complex hierarchies (2 apps, 3 locals, 5 routines) |
| **TestGlobalManager_BeforeInit**                 | Tests error handling before initialization               |

**Key Features Tested:**

* Singleton initialization
* Hierarchical aggregation (Global → App → Local → Routine)
* Count operations without memory allocation
* Error handling for uninitialized state

---

### AppManager Tests (9)

| Test                                              | Description                                       |
| ------------------------------------------------- | ------------------------------------------------- |
| **TestAppManager_CreateApp**                | Tests basic app creation                          |
| **TestAppManager_CreateApp_Idempotent**     | Verifies duplicate creation returns same instance |
| **TestAppManager_CreateApp_AutoInitGlobal** | Tests automatic global manager initialization     |
| **TestAppManager_GetAllLocalManagers**      | Tests local manager listing                       |
| **TestAppManager_GetLocalManagerCount**     | Tests local manager counting                      |
| **TestAppManager_GetGoroutineCount**        | Tests goroutine counting across locals            |
| **TestAppManager_GetAllGoroutines**         | Tests goroutine aggregation                       |
| **TestAppManager_BeforeCreateApp**          | Tests error handling before app creation          |
| **TestAppManager_MultipleApps**             | Tests multiple independent apps                   |

**Key Features Tested:**

* App creation with auto-initialization
* Idempotent operations
* Local manager management
* Goroutine aggregation across local managers
* Multi-app independence

---

### LocalManager Tests (10)

| Test                                              | Description                                       |
| ------------------------------------------------- | ------------------------------------------------- |
| **TestLocalManager_CreateLocal**            | Tests basic local manager creation                |
| **TestLocalManager_CreateLocal_Idempotent** | Verifies duplicate creation returns same instance |
| **TestLocalManager_Go_SpawnGoroutine**      | Tests goroutine spawning and tracking             |
| **TestLocalManager_Go_MultipleGoroutines**  | Tests spawning multiple goroutines                |
| **TestLocalManager_GetAllGoroutines**       | Tests goroutine listing with metadata             |
| **TestLocalManager_GetGoroutineCount**      | Tests goroutine counting                          |
| **TestLocalManager_ContextPropagation**     | Verifies context is passed to goroutines          |
| **TestLocalManager_GoroutineCompletion**    | Tests goroutine completion tracking               |
| **TestLocalManager_MultipleLocalManagers**  | Tests multiple independent local managers         |
| **TestLocalManager_ErrorHandling**          | Tests error propagation from goroutines           |

**Key Features Tested:**

* Local manager creation
* Goroutine spawning with

  Go()
* Context propagation to worker functions
* Goroutine tracking and metadata (ID, function name, start time)
* Completion detection
* Error handling
* Multi-local independence

---

## Bug Fixes During Testing

### Critical Bug: Nil Pointer Dereference

**File:**

types/Intialize.go

**Issue:** The

App(),

Local(), and

Routine() methods accessed `Global.AppManagers` without checking if

Global was nil, causing panics when called before global initialization.

**Fix:** Added nil checks at the start of each method:

<pre><div node="[object Object]" class="relative whitespace-pre-wrap word-break-all p-3 my-2 rounded-sm bg-list-hover-subtle"><div><div class="code-block"><div class="code-line" data-line-number="1"><div class="line-content"><span class="mtk6">func</span><span class="mtk1">(</span><span class="mtk10">Is </span><span class="mtk17">Initializer</span><span class="mtk1">) </span><span class="mtk16">App</span><span class="mtk1">(</span><span class="mtk10">appName</span><span class="mtk1"></span><span class="mtk17">string</span><span class="mtk1">) </span><span class="mtk17">bool</span><span class="mtk1"> {</span></div></div><div class="code-line" data-line-number="2"><div class="line-content"><span class="mtk1"></span><span class="mtk18">if</span><span class="mtk1"></span><span class="mtk10">Global</span><span class="mtk1"></span><span class="mtk3">==</span><span class="mtk1"></span><span class="mtk6">nil</span><span class="mtk1"> {</span></div></div><div class="code-line" data-line-number="3"><div class="line-content"><span class="mtk1"></span><span class="mtk18">return</span><span class="mtk1"></span><span class="mtk6">false</span></div></div><div class="code-line" data-line-number="4"><div class="line-content"><span class="mtk1">    }</span></div></div><div class="code-line" data-line-number="5"><div class="line-content"><span class="mtk1"></span><span class="mtk18">return</span><span class="mtk1"></span><span class="mtk10">Global</span><span class="mtk1">.</span><span class="mtk10">AppManagers</span><span class="mtk1">[</span><span class="mtk10">appName</span><span class="mtk1">] </span><span class="mtk3">!=</span><span class="mtk1"></span><span class="mtk6">nil</span></div></div><div class="code-line" data-line-number="6"><div class="line-content"><span class="mtk1">}</span></div></div><div class="code-line" data-line-number="7"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="8"><div class="line-content"><span class="mtk6">func</span><span class="mtk1">(</span><span class="mtk10">Is </span><span class="mtk17">Initializer</span><span class="mtk1">) </span><span class="mtk16">Local</span><span class="mtk1">(</span><span class="mtk10">appName</span><span class="mtk1">, </span><span class="mtk10">localName</span><span class="mtk1"></span><span class="mtk17">string</span><span class="mtk1">) </span><span class="mtk17">bool</span><span class="mtk1"> {</span></div></div><div class="code-line" data-line-number="9"><div class="line-content"><span class="mtk1"></span><span class="mtk18">if</span><span class="mtk1"></span><span class="mtk10">Global</span><span class="mtk1"></span><span class="mtk3">==</span><span class="mtk1"></span><span class="mtk6">nil</span><span class="mtk1"> {</span></div></div><div class="code-line" data-line-number="10"><div class="line-content"><span class="mtk1"></span><span class="mtk18">return</span><span class="mtk1"></span><span class="mtk6">false</span></div></div><div class="code-line" data-line-number="11"><div class="line-content"><span class="mtk1">    }</span></div></div><div class="code-line" data-line-number="12"><div class="line-content"><span class="mtk1"></span><span class="mtk5">// ... rest of logic</span></div></div><div class="code-line" data-line-number="13"><div class="line-content"><span class="mtk1">}</span></div></div><div class="code-line" data-line-number="14"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="15"><div class="line-content"><span class="mtk6">func</span><span class="mtk1">(</span><span class="mtk10">Is </span><span class="mtk17">Initializer</span><span class="mtk1">) </span><span class="mtk16">Routine</span><span class="mtk1">(</span><span class="mtk10">appName</span><span class="mtk1">, </span><span class="mtk10">localName</span><span class="mtk1">, </span><span class="mtk10">routineID</span><span class="mtk1"></span><span class="mtk17">string</span><span class="mtk1">) </span><span class="mtk17">bool</span><span class="mtk1"> {</span></div></div><div class="code-line" data-line-number="16"><div class="line-content"><span class="mtk1"></span><span class="mtk18">if</span><span class="mtk1"></span><span class="mtk10">Global</span><span class="mtk1"></span><span class="mtk3">==</span><span class="mtk1"></span><span class="mtk6">nil</span><span class="mtk1"> {</span></div></div><div class="code-line" data-line-number="17"><div class="line-content"><span class="mtk1"></span><span class="mtk18">return</span><span class="mtk1"></span><span class="mtk6">false</span></div></div><div class="code-line" data-line-number="18"><div class="line-content"><span class="mtk1">    }</span></div></div><div class="code-line" data-line-number="19"><div class="line-content"><span class="mtk1"></span><span class="mtk5">// ... rest of logic</span></div></div><div class="code-line" data-line-number="20"><div class="line-content"><span class="mtk1">}</span></div></div></div></div></div></pre>

---

## Test Patterns

### Setup Pattern

All tests use a

resetGlobalState() helper to ensure test isolation:

<pre><div node="[object Object]" class="relative whitespace-pre-wrap word-break-all p-3 my-2 rounded-sm bg-list-hover-subtle"><div><div class="code-block"><div class="code-line" data-line-number="1"><div class="line-content"><span class="mtk6">func</span><span class="mtk1"></span><span class="mtk16">resetGlobalState</span><span class="mtk1">() {</span></div></div><div class="code-line" data-line-number="2"><div class="line-content"><span class="mtk1"></span><span class="mtk10">types</span><span class="mtk1">.</span><span class="mtk10">Global</span><span class="mtk1"></span><span class="mtk3">=</span><span class="mtk1"></span><span class="mtk6">nil</span></div></div><div class="code-line" data-line-number="3"><div class="line-content"><span class="mtk1">}</span></div></div></div></div></div></pre>

Each test calls this at the start to prevent state leakage between tests.

### Hierarchy Creation Pattern

Tests create the full hierarchy using the public API:

<pre><div node="[object Object]" class="relative whitespace-pre-wrap word-break-all p-3 my-2 rounded-sm bg-list-hover-subtle"><div><div class="code-block"><div class="code-line" data-line-number="1"><div class="line-content"><span class="mtk5">// Global → App → Local → Routine</span></div></div><div class="code-line" data-line-number="2"><div class="line-content"><span class="mtk10">gm</span><span class="mtk1"></span><span class="mtk3">:=</span><span class="mtk1"></span><span class="mtk10">Global</span><span class="mtk1">.</span><span class="mtk16">NewGlobalManager</span><span class="mtk1">()</span></div></div><div class="code-line" data-line-number="3"><div class="line-content"><span class="mtk10">gm</span><span class="mtk1">.</span><span class="mtk16">Init</span><span class="mtk1">()</span></div></div><div class="code-line" data-line-number="4"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="5"><div class="line-content"><span class="mtk10">appMgr</span><span class="mtk1"></span><span class="mtk3">:=</span><span class="mtk1"></span><span class="mtk10">App</span><span class="mtk1">.</span><span class="mtk16">NewAppManager</span><span class="mtk1">(</span><span class="mtk12">"test-app"</span><span class="mtk1">)</span></div></div><div class="code-line" data-line-number="6"><div class="line-content"><span class="mtk10">app</span><span class="mtk1">, </span><span class="mtk10">_</span><span class="mtk1"></span><span class="mtk3">:=</span><span class="mtk1"></span><span class="mtk10">appMgr</span><span class="mtk1">.</span><span class="mtk16">CreateApp</span><span class="mtk1">()</span></div></div><div class="code-line" data-line-number="7"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="8"><div class="line-content"><span class="mtk10">localMgr</span><span class="mtk1"></span><span class="mtk3">:=</span><span class="mtk1"></span><span class="mtk10">Local</span><span class="mtk1">.</span><span class="mtk16">NewLocalManager</span><span class="mtk1">(</span><span class="mtk12">"test-app"</span><span class="mtk1">, </span><span class="mtk12">"test-local"</span><span class="mtk1">)</span></div></div><div class="code-line" data-line-number="9"><div class="line-content"><span class="mtk10">local</span><span class="mtk1">, </span><span class="mtk10">_</span><span class="mtk1"></span><span class="mtk3">:=</span><span class="mtk1"></span><span class="mtk10">localMgr</span><span class="mtk1">.</span><span class="mtk16">CreateLocal</span><span class="mtk1">(</span><span class="mtk12">"test-local"</span><span class="mtk1">)</span></div></div><div class="code-line" data-line-number="10"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="11"><div class="line-content"><span class="mtk10">app</span><span class="mtk1">.</span><span class="mtk16">AddLocalManager</span><span class="mtk1">(</span><span class="mtk12">"test-local"</span><span class="mtk1">, </span><span class="mtk10">local</span><span class="mtk1">)</span></div></div><div class="code-line" data-line-number="12"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="13"><div class="line-content"><span class="mtk10">routine</span><span class="mtk1"></span><span class="mtk3">:=</span><span class="mtk1"></span><span class="mtk10">types</span><span class="mtk1">.</span><span class="mtk16">NewGoRoutine</span><span class="mtk1">(</span><span class="mtk12">"func1"</span><span class="mtk1">)</span></div></div><div class="code-line" data-line-number="14"><div class="line-content"><span class="mtk10">local</span><span class="mtk1">.</span><span class="mtk16">AddRoutine</span><span class="mtk1">(</span><span class="mtk10">routine</span><span class="mtk1">)</span></div></div></div></div></div></pre>

### Goroutine Testing Pattern

Tests use `atomic.Bool` for thread-safe verification:

<pre><div node="[object Object]" class="relative whitespace-pre-wrap word-break-all p-3 my-2 rounded-sm bg-list-hover-subtle"><div><div class="code-block"><div class="code-line" data-line-number="1"><div class="line-content"><span class="mtk10">executed</span><span class="mtk1"></span><span class="mtk3">:=</span><span class="mtk1"></span><span class="mtk17">atomic</span><span class="mtk1">.</span><span class="mtk17">Bool</span><span class="mtk1">{}</span></div></div><div class="code-line" data-line-number="2"><div class="line-content"><span class="mtk10">localMgr</span><span class="mtk1">.</span><span class="mtk16">Go</span><span class="mtk1">(</span><span class="mtk12">"testFunc"</span><span class="mtk1">, </span><span class="mtk6">func</span><span class="mtk1">(</span><span class="mtk10">ctx</span><span class="mtk1"></span><span class="mtk17">context</span><span class="mtk1">.</span><span class="mtk17">Context</span><span class="mtk1">) </span><span class="mtk17">error</span><span class="mtk1"> {</span></div></div><div class="code-line" data-line-number="3"><div class="line-content"><span class="mtk1"></span><span class="mtk10">executed</span><span class="mtk1">.</span><span class="mtk16">Store</span><span class="mtk1">(</span><span class="mtk6">true</span><span class="mtk1">)</span></div></div><div class="code-line" data-line-number="4"><div class="line-content"><span class="mtk1"></span><span class="mtk18">return</span><span class="mtk1"></span><span class="mtk6">nil</span></div></div><div class="code-line" data-line-number="5"><div class="line-content"><span class="mtk1">})</span></div></div><div class="code-line" data-line-number="6"><div class="line-content"><span class="mtk1"></span></div></div><div class="code-line" data-line-number="7"><div class="line-content"><span class="mtk10">time</span><span class="mtk1">.</span><span class="mtk16">Sleep</span><span class="mtk1">(</span><span class="mtk7">100</span><span class="mtk1"></span><span class="mtk3">*</span><span class="mtk1"></span><span class="mtk10">time</span><span class="mtk1">.</span><span class="mtk10">Millisecond</span><span class="mtk1">)</span></div></div><div class="code-line" data-line-number="8"><div class="line-content"><span class="mtk18">if</span><span class="mtk1"></span><span class="mtk3">!</span><span class="mtk10">executed</span><span class="mtk1">.</span><span class="mtk16">Load</span><span class="mtk1">() {</span></div></div><div class="code-line" data-line-number="9"><div class="line-content"><span class="mtk1"></span><span class="mtk10">t</span><span class="mtk1">.</span><span class="mtk16">Error</span><span class="mtk1">(</span><span class="mtk12">"Goroutine was not executed"</span><span class="mtk1">)</span></div></div><div class="code-line" data-line-number="10"><div class="line-content"><span class="mtk1">}</span></div></div></div></div></div></pre>

---

## Test Results

<pre><div node="[object Object]" class="relative whitespace-pre-wrap word-break-all p-3 my-2 rounded-sm bg-list-hover-subtle"><div><div class="code-block"><div class="code-line" data-line-number="1"><div class="line-content"><span class="mtk1">=== RUN   TestAppManager_CreateApp</span></div></div><div class="code-line" data-line-number="2"><div class="line-content"><span class="mtk1">--- PASS: TestAppManager_CreateApp (0.00s)</span></div></div><div class="code-line" data-line-number="3"><div class="line-content"><span class="mtk1">=== RUN   TestAppManager_CreateApp_Idempotent</span></div></div><div class="code-line" data-line-number="4"><div class="line-content"><span class="mtk1">--- PASS: TestAppManager_CreateApp_Idempotent (0.00s)</span></div></div><div class="code-line" data-line-number="5"><div class="line-content"><span class="mtk1">=== RUN   TestAppManager_CreateApp_AutoInitGlobal</span></div></div><div class="code-line" data-line-number="6"><div class="line-content"><span class="mtk1">--- PASS: TestAppManager_CreateApp_AutoInitGlobal (0.00s)</span></div></div><div class="code-line" data-line-number="7"><div class="line-content"><span class="mtk1">=== RUN   TestAppManager_GetAllLocalManagers</span></div></div><div class="code-line" data-line-number="8"><div class="line-content"><span class="mtk1">--- PASS: TestAppManager_GetAllLocalManagers (0.00s)</span></div></div><div class="code-line" data-line-number="9"><div class="line-content"><span class="mtk1">=== RUN   TestAppManager_GetLocalManagerCount</span></div></div><div class="code-line" data-line-number="10"><div class="line-content"><span class="mtk1">--- PASS: TestAppManager_GetLocalManagerCount (0.00s)</span></div></div><div class="code-line" data-line-number="11"><div class="line-content"><span class="mtk1">=== RUN   TestAppManager_GetGoroutineCount</span></div></div><div class="code-line" data-line-number="12"><div class="line-content"><span class="mtk1">--- PASS: TestAppManager_GetGoroutineCount (0.00s)</span></div></div><div class="code-line" data-line-number="13"><div class="line-content"><span class="mtk1">=== RUN   TestAppManager_GetAllGoroutines</span></div></div><div class="code-line" data-line-number="14"><div class="line-content"><span class="mtk1">--- PASS: TestAppManager_GetAllGoroutines (0.00s)</span></div></div><div class="code-line" data-line-number="15"><div class="line-content"><span class="mtk1">=== RUN   TestAppManager_BeforeCreateApp</span></div></div><div class="code-line" data-line-number="16"><div class="line-content"><span class="mtk1">--- PASS: TestAppManager_BeforeCreateApp (0.00s)</span></div></div><div class="code-line" data-line-number="17"><div class="line-content"><span class="mtk1">=== RUN   TestAppManager_MultipleApps</span></div></div><div class="code-line" data-line-number="18"><div class="line-content"><span class="mtk1">--- PASS: TestAppManager_MultipleApps (0.00s)</span></div></div><div class="code-line" data-line-number="19"><div class="line-content"><span class="mtk1">=== RUN   TestGlobalManager_Init</span></div></div><div class="code-line" data-line-number="20"><div class="line-content"><span class="mtk1">--- PASS: TestGlobalManager_Init (0.00s)</span></div></div><div class="code-line" data-line-number="21"><div class="line-content"><span class="mtk1">=== RUN   TestGlobalManager_GetAllAppManagers_Empty</span></div></div><div class="code-line" data-line-number="22"><div class="line-content"><span class="mtk1">--- PASS: TestGlobalManager_GetAllAppManagers_Empty (0.00s)</span></div></div><div class="code-line" data-line-number="23"><div class="line-content"><span class="mtk1">=== RUN   TestGlobalManager_GetAppManagerCount</span></div></div><div class="code-line" data-line-number="24"><div class="line-content"><span class="mtk1">--- PASS: TestGlobalManager_GetAppManagerCount (0.00s)</span></div></div><div class="code-line" data-line-number="25"><div class="line-content"><span class="mtk1">=== RUN   TestGlobalManager_GetAllAppManagers_WithApps</span></div></div><div class="code-line" data-line-number="26"><div class="line-content"><span class="mtk1">--- PASS: TestGlobalManager_GetAllAppManagers_WithApps (0.00s)</span></div></div><div class="code-line" data-line-number="27"><div class="line-content"><span class="mtk1">=== RUN   TestGlobalManager_GetAllLocalManagers</span></div></div><div class="code-line" data-line-number="28"><div class="line-content"><span class="mtk1">--- PASS: TestGlobalManager_GetAllLocalManagers (0.00s)</span></div></div><div class="code-line" data-line-number="29"><div class="line-content"><span class="mtk1">=== RUN   TestGlobalManager_GetLocalManagerCount</span></div></div><div class="code-line" data-line-number="30"><div class="line-content"><span class="mtk1">--- PASS: TestGlobalManager_GetLocalManagerCount (0.00s)</span></div></div><div class="code-line" data-line-number="31"><div class="line-content"><span class="mtk1">=== RUN   TestGlobalManager_GetAllGoroutines</span></div></div><div class="code-line" data-line-number="32"><div class="line-content"><span class="mtk1">--- PASS: TestGlobalManager_GetAllGoroutines (0.00s)</span></div></div><div class="code-line" data-line-number="33"><div class="line-content"><span class="mtk1">=== RUN   TestGlobalManager_GetGoroutineCount</span></div></div><div class="code-line" data-line-number="34"><div class="line-content"><span class="mtk1">--- PASS: TestGlobalManager_GetGoroutineCount (0.00s)</span></div></div><div class="code-line" data-line-number="35"><div class="line-content"><span class="mtk1">=== RUN   TestGlobalManager_MultipleAppsAndLocals</span></div></div><div class="code-line" data-line-number="36"><div class="line-content"><span class="mtk1">--- PASS: TestGlobalManager_MultipleAppsAndLocals (0.00s)</span></div></div><div class="code-line" data-line-number="37"><div class="line-content"><span class="mtk1">=== RUN   TestGlobalManager_BeforeInit</span></div></div><div class="code-line" data-line-number="38"><div class="line-content"><span class="mtk1">--- PASS: TestGlobalManager_BeforeInit (0.00s)</span></div></div><div class="code-line" data-line-number="39"><div class="line-content"><span class="mtk1">=== RUN   TestLocalManager_CreateLocal</span></div></div><div class="code-line" data-line-number="40"><div class="line-content"><span class="mtk1">--- PASS: TestLocalManager_CreateLocal (0.00s)</span></div></div><div class="code-line" data-line-number="41"><div class="line-content"><span class="mtk1">=== RUN   TestLocalManager_CreateLocal_Idempotent</span></div></div><div class="code-line" data-line-number="42"><div class="line-content"><span class="mtk1">--- PASS: TestLocalManager_CreateLocal_Idempotent (0.00s)</span></div></div><div class="code-line" data-line-number="43"><div class="line-content"><span class="mtk1">=== RUN   TestLocalManager_Go_SpawnGoroutine</span></div></div><div class="code-line" data-line-number="44"><div class="line-content"><span class="mtk1">--- PASS: TestLocalManager_Go_SpawnGoroutine (0.10s)</span></div></div><div class="code-line" data-line-number="45"><div class="line-content"><span class="mtk1">=== RUN   TestLocalManager_Go_MultipleGoroutines</span></div></div><div class="code-line" data-line-number="46"><div class="line-content"><span class="mtk1">--- PASS: TestLocalManager_Go_MultipleGoroutines (0.00s)</span></div></div><div class="code-line" data-line-number="47"><div class="line-content"><span class="mtk1">=== RUN   TestLocalManager_GetAllGoroutines</span></div></div><div class="code-line" data-line-number="48"><div class="line-content"><span class="mtk1">--- PASS: TestLocalManager_GetAllGoroutines (0.00s)</span></div></div><div class="code-line" data-line-number="49"><div class="line-content"><span class="mtk1">=== RUN   TestLocalManager_GetGoroutineCount</span></div></div><div class="code-line" data-line-number="50"><div class="line-content"><span class="mtk1">--- PASS: TestLocalManager_GetGoroutineCount (0.00s)</span></div></div><div class="code-line" data-line-number="51"><div class="line-content"><span class="mtk1">=== RUN   TestLocalManager_ContextPropagation</span></div></div><div class="code-line" data-line-number="52"><div class="line-content"><span class="mtk1">--- PASS: TestLocalManager_ContextPropagation (0.10s)</span></div></div><div class="code-line" data-line-number="53"><div class="line-content"><span class="mtk1">=== RUN   TestLocalManager_GoroutineCompletion</span></div></div><div class="code-line" data-line-number="54"><div class="line-content"><span class="mtk1">--- PASS: TestLocalManager_GoroutineCompletion (0.15s)</span></div></div><div class="code-line" data-line-number="55"><div class="line-content"><span class="mtk1">=== RUN   TestLocalManager_MultipleLocalManagers</span></div></div><div class="code-line" data-line-number="56"><div class="line-content"><span class="mtk1">--- PASS: TestLocalManager_MultipleLocalManagers (0.00s)</span></div></div><div class="code-line" data-line-number="57"><div class="line-content"><span class="mtk1">=== RUN   TestLocalManager_ErrorHandling</span></div></div><div class="code-line" data-line-number="58"><div class="line-content"><span class="mtk1">--- PASS: TestLocalManager_ErrorHandling (0.10s)</span></div></div><div class="code-line" data-line-number="59"><div class="line-content"><span class="mtk1">PASS</span></div></div><div class="code-line" data-line-number="60"><div class="line-content"><span class="mtk1">ok      github.com/neerajchowdary889/GoRoutinesManager/Tests    0.455s</span></div></div></div></div></div></pre>

---

## Future Test Additions

Potential areas for additional testing:

1. **Shutdown Tests** - Once shutdown logic is implemented
2. **Function-level Wait Groups** - Test

   NewFunctionWaitGroup() when implemented
3. **Concurrent Operations** - Stress testing with concurrent goroutine spawning
4. **Context Cancellation** - Test cascading cancellation through hierarchy
5. **Performance Tests** - Benchmark goroutine spawning and tracking overhead
6. **Race Condition Tests** - Run with `-race` flag
7. **Memory Leak Tests** - Verify goroutines are properly cleaned up

---

## Notes

* All tests use the public API (interfaces) rather than accessing internal types directly
* Tests verify both success and error paths
* Goroutine tests include timing delays to ensure async operations complete
* Tests are isolated using

  resetGlobalState() to prevent interference
* The test suite discovered and fixed a critical nil pointer dereference bug
