/** The global namespace for the app */
declare namespace App {
  /** Theme namespace */
  namespace Theme {
    type ColorPaletteNumber = import('@sa/color').ColorPaletteNumber;

    /** Theme setting */
    interface ThemeSetting {
      /** Theme scheme */
      themeScheme: UnionKey.ThemeScheme;
      /** grayscale mode */
      grayscale: boolean;
      /** colour weakness mode */
      colourWeakness: boolean;
      /** Whether to recommend color */
      recommendColor: boolean;
      /** Theme color */
      themeColor: string;
      /** Other color */
      otherColor: OtherColor;
      /** Whether info color is followed by the primary color */
      isInfoFollowPrimary: boolean;
      /** Layout */
      layout: {
        /** Layout mode */
        mode: UnionKey.ThemeLayoutMode;
        /** Scroll mode */
        scrollMode: UnionKey.ThemeScrollMode;
        /**
         * Whether to reverse the horizontal mix
         *
         * if true, the vertical child level menus in left and horizontal first level menus in top
         */
        reverseHorizontalMix?: boolean;
      };
      /** Page */
      page: {
        /** Whether to show the page transition */
        animate: boolean;
        /** Page animate mode */
        animateMode: UnionKey.ThemePageAnimateMode;
      };
      /** Header */
      header: {
        /** Header height */
        height: number;
        /** Header breadcrumb */
        breadcrumb: {
          /** Whether to show the breadcrumb */
          visible: boolean;
          /** Whether to show the breadcrumb icon */
          showIcon: boolean;
        };
      };
      /** Tab */
      tab: {
        /** Whether to show the tab */
        visible: boolean;
        /**
         * Whether to cache the tab
         *
         * If cache, the tabs will get from the local storage when the page is refreshed
         */
        cache: boolean;
        /** Tab height */
        height: number;
        /** Tab mode */
        mode: UnionKey.ThemeTabMode;
      };
      /** Fixed header and tab */
      fixedHeaderAndTab: boolean;
      /** Sider */
      sider: {
        /** Inverted sider */
        inverted: boolean;
        /** Sider width */
        width: number;
        /** Collapsed sider width */
        collapsedWidth: number;
        /** Sider width when the layout is 'vertical-mix' or 'horizontal-mix' */
        mixWidth: number;
        /** Collapsed sider width when the layout is 'vertical-mix' or 'horizontal-mix' */
        mixCollapsedWidth: number;
        /** Child menu width when the layout is 'vertical-mix' or 'horizontal-mix' */
        mixChildMenuWidth: number;
      };
      /** Footer */
      footer: {
        /** Whether to show the footer */
        visible: boolean;
        /** Whether fixed the footer */
        fixed: boolean;
        /** Footer height */
        height: number;
        /** Whether float the footer to the right when the layout is 'horizontal-mix' */
        right: boolean;
      };
      /** Watermark */
      watermark?: {
        /** Whether to show the watermark */
        visible: boolean;
        /** Watermark text */
        text: string;
      };
      /** define some theme settings tokens, will transform to css variables */
      tokens: {
        light: ThemeSettingToken;
        dark?: {
          [K in keyof ThemeSettingToken]?: Partial<ThemeSettingToken[K]>;
        };
      };
    }

    interface OtherColor {
      info: string;
      success: string;
      warning: string;
      error: string;
    }

    interface ThemeColor extends OtherColor {
      primary: string;
    }

    type ThemeColorKey = keyof ThemeColor;

    type ThemePaletteColor = {
      [key in ThemeColorKey | `${ThemeColorKey}-${ColorPaletteNumber}`]: string;
    };

    type BaseToken = Record<string, Record<string, string>>;

    interface ThemeSettingTokenColor {
      /** the progress bar color, if not set, will use the primary color */
      nprogress?: string;
      container: string;
      layout: string;
      inverted: string;
      'base-text': string;
    }

    interface ThemeSettingTokenBoxShadow {
      header: string;
      sider: string;
      tab: string;
    }

    interface ThemeSettingToken {
      colors: ThemeSettingTokenColor;
      boxShadow: ThemeSettingTokenBoxShadow;
    }

    type ThemeTokenColor = ThemePaletteColor & ThemeSettingTokenColor;

    /** Theme token CSS variables */
    type ThemeTokenCSSVars = {
      colors: ThemeTokenColor & { [key: string]: string };
      boxShadow: ThemeSettingTokenBoxShadow & { [key: string]: string };
    };
  }

  /** Global namespace */
  namespace Global {
    type VNode = import('vue').VNode;
    type RouteLocationNormalizedLoaded = import('vue-router').RouteLocationNormalizedLoaded;
    type RouteKey = import('@elegant-router/types').RouteKey;
    type RouteMap = import('@elegant-router/types').RouteMap;
    type RoutePath = import('@elegant-router/types').RoutePath;
    type LastLevelRouteKey = import('@elegant-router/types').LastLevelRouteKey;

    /** The global header props */
    interface HeaderProps {
      /** Whether to show the logo */
      showLogo?: boolean;
      /** Whether to show the menu toggler */
      showMenuToggler?: boolean;
      /** Whether to show the menu */
      showMenu?: boolean;
    }

    /** The global menu */
    type Menu = {
      /**
       * The menu key
       *
       * Equal to the route key
       */
      key: string;
      /** The menu label */
      label: string;
      /** The menu i18n key */
      i18nKey?: I18n.I18nKey | null;
      /** The route key */
      routeKey: RouteKey;
      /** The route path */
      routePath: RoutePath;
      /** The menu icon */
      icon?: () => VNode;
      /** The menu children */
      children?: Menu[];
    };

    type Breadcrumb = Omit<Menu, 'children'> & {
      options?: Breadcrumb[];
    };

    /** Tab route */
    type TabRoute = Pick<RouteLocationNormalizedLoaded, 'name' | 'path' | 'meta'> &
      Partial<Pick<RouteLocationNormalizedLoaded, 'fullPath' | 'query' | 'matched'>>;

    /** The global tab */
    type Tab = {
      /** The tab id */
      id: string;
      /** The tab label */
      label: string;
      /**
       * The new tab label
       *
       * If set, the tab label will be replaced by this value
       */
      newLabel?: string;
      /**
       * The old tab label
       *
       * when reset the tab label, the tab label will be replaced by this value
       */
      oldLabel?: string;
      /** The tab route key */
      routeKey: LastLevelRouteKey;
      /** The tab route path */
      routePath: RouteMap[LastLevelRouteKey];
      /** The tab route full path */
      fullPath: string;
      /** The tab fixed index */
      fixedIndex?: number | null;
      /**
       * Tab icon
       *
       * Iconify icon
       */
      icon?: string;
      /**
       * Tab local icon
       *
       * Local icon
       */
      localIcon?: string;
      /** I18n key */
      i18nKey?: I18n.I18nKey | null;
    };

    /** Form rule */
    type FormRule = import('naive-ui').FormItemRule;

    /** The global dropdown key */
    type DropdownKey = 'closeCurrent' | 'closeOther' | 'closeLeft' | 'closeRight' | 'closeAll';
  }

  /**
   * I18n namespace
   *
   * Locales type
   */
  namespace I18n {
    type RouteKey = import('@elegant-router/types').RouteKey;

    type LangType = 'en-US' | 'zh-CN';

    type LangOption = {
      label: string;
      key: LangType;
    };

    type I18nRouteKey = Exclude<RouteKey, 'root' | 'not-found'>;

    type FormMsg = {
      required: string;
      invalid: string;
    };

    type Schema = {
      system: {
        title: string;
        updateTitle: string;
        updateContent: string;
        updateConfirm: string;
        updateCancel: string;
      };
      common: {
        action: string;
        add: string;
        addSuccess: string;
        backToHome: string;
        batchDelete: string;
        cancel: string;
        close: string;
        check: string;
        expandColumn: string;
        columnSetting: string;
        config: string;
        confirm: string;
        delete: string;
        deleteSuccess: string;
        confirmDelete: string;
        download: string;
        edit: string;
        editSuccess: string;
        editHomeSuccess: string;
        warning: string;
        error: string;
        index: string;
        keywordSearch: string;
        logout: string;
        logoutConfirm: string;
        lookForward: string;
        modify: string;
        modifySuccess: string;
        noData: string;
        operate: string;
        pleaseCheckValue: string;
        refresh: string;
        reset: string;
        search: string;
        switch: string;
        tip: string;
        trigger: string;
        update: string;
        updateSuccess: string;
        userCenter: string;
        yesOrNo: {
          yes: string;
          no: string;
        };
      };
      request: {
        logout: string;
        logoutMsg: string;
        logoutWithModal: string;
        logoutWithModalMsg: string;
        refreshToken: string;
        tokenExpired: string;
      };
      theme: {
        themeSchema: { title: string } & Record<UnionKey.ThemeScheme, string>;
        grayscale: string;
        colourWeakness: string;
        layoutMode: { title: string; reverseHorizontalMix: string } & Record<UnionKey.ThemeLayoutMode, string>;
        recommendColor: string;
        recommendColorDesc: string;
        themeColor: {
          title: string;
          followPrimary: string;
        } & Theme.ThemeColor;
        scrollMode: { title: string } & Record<UnionKey.ThemeScrollMode, string>;
        page: {
          animate: string;
          mode: { title: string } & Record<UnionKey.ThemePageAnimateMode, string>;
        };
        fixedHeaderAndTab: string;
        header: {
          height: string;
          breadcrumb: {
            visible: string;
            showIcon: string;
          };
        };
        tab: {
          visible: string;
          cache: string;
          height: string;
          mode: { title: string } & Record<UnionKey.ThemeTabMode, string>;
        };
        sider: {
          inverted: string;
          width: string;
          collapsedWidth: string;
          mixWidth: string;
          mixCollapsedWidth: string;
          mixChildMenuWidth: string;
        };
        footer: {
          visible: string;
          fixed: string;
          height: string;
          right: string;
        };
        watermark: {
          visible: string;
          text: string;
        };
        themeDrawerTitle: string;
        pageFunTitle: string;
        configOperation: {
          copyConfig: string;
          copySuccessMsg: string;
          resetConfig: string;
          resetSuccessMsg: string;
        };
      };
      route: Record<I18nRouteKey, string>;
      button: {
        manage: {
          user: {
            list: string;
            add: string;
            delete: string;
            edit: string;
          };
          menu: {
            list: string;
            add: string;
            delete: string;
            edit: string;
          };
          role: {
            list: string;
            add: string;
            delete: string;
            edit: string;
          };
        };
        hrm: {
          org: {
            add: string;
            addchild: string;
            edit: string;
            delete: string;
            import: string;
            export: string;
          };
          shift: {
            add: string;
            edit: string;
            delete: string;
          };
          leavetype: {
            add: string;
            edit: string;
            delete: string;
          };
          leavebalance: {
            init: string;
            adjust: string;
          };
          attendancegroup: {
            add: string;
            edit: string;
            delete: string;
            togglestatus: string;
            members: string;
            shiftusage: string;
          };
          attendanceconfig: {
            save: string;
            delete: string;
          };
          punch: {
            punch: string;
            myList: string;
            add: string;
            edit: string;
            delete: string;
            export: string;
            import: string;
            settle: string;
          };
          overtimerule: {
            add: string;
            edit: string;
            delete: string;
            togglestatus: string;
          };
          schedulecycle: {
            add: string;
            edit: string;
            delete: string;
            togglestatus: string;
          };
          scheduleboard: {
            savecell: string;
            batchsave: string;
            generate: string;
            copyweek: string;
            import: string;
            export: string;
            publish: string;
          };
          transfer: {
            add: string;
            edit: string;
            delete: string;
          };
          regularization: {
            add: string;
            edit: string;
            delete: string;
          };
          resignation: {
            add: string;
            edit: string;
            delete: string;
          };
          entry: {
            add: string;
            edit: string;
            delete: string;
          };
          reward: {
            add: string;
            edit: string;
            delete: string;
          };
          requirement: {
            add: string;
            edit: string;
            delete: string;
          };
          candidate: {
            add: string;
            edit: string;
            delete: string;
            import: string;
            updatestatus: string;
          };
          bill: {
            submit: string;
            approve: string;
            reject: string;
            revoke: string;
          };
          training: {
            downloadCert: string;
            plan: {
              add: string;
              edit: string;
              delete: string;
            };
            session: {
              add: string;
              edit: string;
              delete: string;
              participants: string;
            };
            certTemplate: {
              add: string;
              edit: string;
              toggleStatus: string;
              delete: string;
            };
            certificate: {
              batchIssue: string;
              import: string;
              issue: string;
              revoke: string;
              uploadFile: string;
            };
          };
          salary_subject: {
            add: string;
            edit: string;
            delete: string;
          };
          salary_item: {
            add: string;
            edit: string;
            delete: string;
          };
          tax_profile: {
            add: string;
            edit: string;
            delete: string;
            export: string;
          };
          tax_special_deduction: {
            maintain: string;
            delete: string;
            maintain_year: string;
            import: string;
            export: string;
          };
          tax_opening: {
            add: string;
            edit: string;
            delete: string;
            import: string;
            export: string;
          };
          employee: {
            add: string;
            edit: string;
            delete: string;
            import: string;
            export: string;
            generateuser: string;
          };
          contract: {
            add: string;
            edit: string;
            delete: string;
            renew: string;
          };
          empcert: {
            add: string;
            edit: string;
            delete: string;
          };
          performance: {
            indicator: {
              add: string;
              edit: string;
              delete: string;
            };
            template: {
              add: string;
              edit: string;
              delete: string;
            };
            scheme: {
              add: string;
              edit: string;
              delete: string;
            };
          };
        };
      };
      page: {
        login: {
          common: {
            loginOrRegister: string;
            usernamePlaceholder: string;
            phonePlaceholder: string;
            emailPlaceholder: string;
            codePlaceholder: string;
            passwordPlaceholder: string;
            confirmPasswordPlaceholder: string;
            codeLogin: string;
            confirm: string;
            back: string;
            registerSuccess: string;
            validateSuccess: string;
            loginSuccess: string;
            welcomeBack: string;
          };
          pwdLogin: {
            title: string;
            rememberMe: string;
            forgetPassword: string;
            register: string;
            otherAccountLogin: string;
            otherLoginMode: string;
            superAdmin: string;
            admin: string;
            user: string;
          };
          codeLogin: {
            title: string;
            emailType: string;
            phoneType: string;
            getCode: string;
            reGetCode: string;
            sendCodeSuccess: string;
            imageCodePlaceholder: string;
          };
          register: {
            title: string;
            agreement: string;
            protocol: string;
            policy: string;
          };
          resetPwd: {
            title: string;
          };
          bindWeChat: {
            title: string;
          };
        };
        about: {
          title: string;
          introduction: string;
          projectInfo: {
            title: string;
            version: string;
            latestBuildTime: string;
            githubLink: string;
            previewLink: string;
          };
          prdDep: string;
          devDep: string;
        };
        home: {
          branchDesc: string;
          greeting: string;
          weatherDesc: string;
          projectNews: {
            title: string;
            moreNews: string;
            desc1: string;
            desc2: string;
            desc3: string;
            desc4: string;
            desc5: string;
          };
          creativity: string;
        };
        manage: {
          common: {
            status: {
              enable: string;
              disable: string;
            };
          };
          role: {
            title: string;
            roleName: string;
            roleCode: string;
            roleStatus: string;
            roleDesc: string;
            form: {
              roleName: string;
              roleCode: string;
              roleStatus: string;
              roleDesc: string;
            };
            addRole: string;
            editRole: string;
            menuAuth: string;
            buttonAuth: string;
          };
          user: {
            title: string;
            username: string;
            userGender: string;
            nickName: string;
            userPhone: string;
            userEmail: string;
            userStatus: string;
            userRole: string;
            password: string;
            form: {
              username: string;
              userGender: string;
              nickName: string;
              userPhone: string;
              userEmail: string;
              userStatus: string;
              userRole: string;
              password: string;
            };
            addUser: string;
            editUser: string;
            gender: {
              male: string;
              female: string;
            };
          };
          menu: {
            home: string;
            title: string;
            id: string;
            parentId: string;
            menuType: string;
            menuName: string;
            routeName: string;
            routePath: string;
            pathParam: string;
            layout: string;
            page: string;
            i18nKey: string;
            icon: string;
            localIcon: string;
            iconTypeTitle: string;
            order: string;
            constant: string;
            keepAlive: string;
            href: string;
            iframePage: string;
            hideInMenu: string;
            activeMenu: string;
            multiTab: string;
            fixedIndexInTab: string;
            query: string;
            button: string;
            buttonCode: string;
            buttonDesc: string;
            permission: string;
            permissionCode: string;
            permissionDesc: string;
            menuStatus: string;
            form: {
              home: string;
              menuType: string;
              menuName: string;
              routeName: string;
              routePath: string;
              pathParam: string;
              layout: string;
              page: string;
              i18nKey: string;
              icon: string;
              localIcon: string;
              order: string;
              keepAlive: string;
              href: string;
              hideInMenu: string;
              activeMenu: string;
              multiTab: string;
              fixedInTab: string;
              fixedIndexInTab: string;
              queryKey: string;
              queryValue: string;
              button: string;
              buttonCode: string;
              buttonDesc: string;
              permissionCode: string;
              permissionDesc: string;
              menuStatus: string;
            };
            addMenu: string;
            editMenu: string;
            addChildMenu: string;
            addButton: string;
            type: {
              directory: string;
              menu: string;
              button: string;
            };
            iconType: {
              iconify: string;
              local: string;
            };
          };
        };
        hrm: {
          performance: {
            common: {
              status: {
                enable: string;
                disable: string;
              };
            };
            indicatorType: {
              kpi: string;
              behavior: string;
            };
            scoreRule: {
              formula: string;
              manual: string;
            };
            dataSource: {
              system: string;
              manual: string;
            };
            schemePeriodType: {
              month: string;
              quarter: string;
              half_year: string;
              year: string;
            };
            schemeAssessmentMode: {
              kpi: string;
              kpi_behavior: string;
            };
            schemeScoreMode: {
              percent: string;
              grade: string;
              amount: string;
            };
            schemePayLinkMode: {
              coefficient: string;
              amount: string;
            };
            indicator: {
              title: string;
              code: string;
              name: string;
              indicatorType: string;
              unit: string;
              scoreRule: string;
              dataSource: string;
              defaultWeight: string;
              status: string;
              remark: string;
              form: {
                code: string;
                name: string;
                indicatorType: string;
                unit: string;
                scoreRule: string;
                dataSource: string;
                status: string;
                remark: string;
              };
              addIndicator: string;
              editIndicator: string;
            };
            template: {
              title: string;
              name: string;
              status: string;
              remark: string;
              createTime: string;
              form: {
                name: string;
                status: string;
                remark: string;
              };
              configItems: string;
              indicatorPlaceholder: string;
              addIndicatorItem: string;
              weightSumLabel: string;
              weightSumError: string;
              itemIncomplete: string;
              addTemplate: string;
              editTemplate: string;
            };
            scheme: {
              title: string;
              name: string;
              templateName: string;
              periodType: string;
              assessmentMode: string;
              scoreMode: string;
              payLinkMode: string;
              status: string;
              remark: string;
              form: {
                name: string;
                templateUuid: string;
                periodType: string;
                assessmentMode: string;
                scoreMode: string;
                payLinkMode: string;
                status: string;
                remark: string;
              };
              addScheme: string;
              editScheme: string;
            };
          };
          org: {
            title: string;
            orgName: string;
            orgTypeTitle: string;
            orgType: {
              company: string;
              department: string;
            };
            leader: string;
            orderNum: string;
            orgStatus: string;
            status: {
              open: string;
              closed: string;
            };
            remark: string;
            createTime: string;
            parentOrg: string;
            form: {
              orgName: string;
              orgType: string;
              leader: string;
              orderNum: string;
              orgStatus: string;
              remark: string;
              parentOrg: string;
            };
            addOrg: string;
            addChildOrg: string;
            editOrg: string;
            importOrg: string;
            exportOrg: string;
            collapse: string;
            expand: string;
            importSuccess: string;
            chooseFileTip: string;
            importModalTitle: string;
            orgChart: {
              title: string;
              memberCount: string;
              memberCountTip: string;
              totalMemberCount: string;
              totalMemberCountTip: string;
              expandAll: string;
              collapseAll: string;
              refresh: string;
              empty: string;
              noLeader: string;
            };
          };
          attendance: {
            yes: string;
            no: string;
            status: {
              enabled: string;
              disabled: string;
            };
            shift: {
              title: string;
              shiftName: string;
              workTime: string;
              crossDay: string;
              nextDay: string;
              restTime: string;
              noRest: string;
              standardHours: string;
              flexibleMinutes: string;
              notFlexible: string;
              flexibleMinutesValue: string;
              status: string;
              remark: string;
              addShift: string;
              editShift: string;
              form: {
                shiftName: string;
                startTime: string;
                endTime: string;
                rest: string;
                flexibleMinutes: string;
                status: string;
                remark: string;
              };
            };
            group: {
              title: string;
              groupName: string;
              groupType: {
                title: string;
                fixed: string;
                scheduling: string;
                flexible: string;
              };
              workTime: string;
              workDays: string;
              scope: string;
              scopeDepts: string;
              scopeMembers: string;
              managers: string;
              status: string;
              remark: string;
              enable: string;
              disable: string;
              confirmEnable: string;
              confirmDisable: string;
              flexibleNoTime: string;
              schedulingFallback: string;
              members: string;
              noMembers: string;
              addGroup: string;
              editGroup: string;
              weekday: {
                1: string;
                2: string;
                3: string;
                4: string;
                5: string;
                6: string;
                7: string;
              };
              weekdayShort: {
                1: string;
                2: string;
                3: string;
                4: string;
                5: string;
                6: string;
                7: string;
              };
              form: {
                groupName: string;
                groupType: string;
                workDays: string;
                orgScope: string;
                memberScope: string;
                managers: string;
                status: string;
                remark: string;
              };
            };
            config: {
              title: string;
              org: string;
              orgName: string;
              quickStart: string;
              editConfig: string;
              recheck: string;
              fallbackHint: string;
              noRule: string;
              sourceGroup: string;
              sourceOrg: string;
              healthEmployee: string;
              healthSource: string;
              healthyTotal: string;
              healthyGroup: string;
              healthyOrg: string;
              healthyNoRule: string;
              form: {
                orgName: string;
                org: string;
              };
            };
            leaveType: {
              title: string;
              typeName: string;
              typeCode: string;
              typeValue: string;
              balanceEnabled: string;
              attachmentRequired: string;
              defaultDays: string;
              orderNum: string;
              status: string;
              remark: string;
              addLeaveType: string;
              editLeaveType: string;
              form: {
                typeName: string;
                typeCode: string;
                typeValue: string;
                status: string;
                remark: string;
              };
            };
            leaveBalance: {
              title: string;
              year: string;
              name: string;
              orgName: string;
              orgUuid: string;
              leaveTypeCode: string;
              employeeNo: string;
              employeeName: string;
              hireDate: string;
              leaveTypeName: string;
              granted: string;
              used: string;
              frozen: string;
              available: string;
              initTitle: string;
              initScope: string;
              initScopeHint: string;
              initSuccess: string;
              defaultDays: string;
              skipped: string;
              skippedYes: string;
              skippedNo: string;
              adjustTitle: string;
              adjustDays: string;
              adjustDaysPlaceholder: string;
              adjustDaysRequired: string;
              adjustSuccess: string;
              reason: string;
              logTitle: string;
              logDays: string;
              logTime: string;
              changeType: string;
              changeTypeInit: string;
              changeTypeAdjust: string;
              changeTypeConsume: string;
              changeTypeAccrue: string;
              operatorName: string;
              billUuid: string;
              form: {
                year: string;
                name: string;
                leaveTypeCode: string;
                reason: string;
              };
            };
            attendancePunch: {
              title: string;
              listTitle: string;
              clockIn: string;
              clockOut: string;
              clockInSuccess: string;
              clockOutSuccess: string;
            };
            punchRecord: {
              title: string;
              dateRange: string;
              name: string;
              employeeName: string;
              username: string;
              punchDate: string;
              weekday: string;
              clockInTime: string;
              clockInStatus: string;
              clockInLocation: string;
              clockOutTime: string;
              clockOutStatus: string;
              clockOutLocation: string;
              overtime: string;
              overtimeYes: string;
              overtimeNo: string;
              employeeUuid: string;
              statusNormal: string;
              statusField: string;
              statusMissing: string;
              settleStatus: string;
              settleNormal: string;
              settleLate: string;
              settleEarly: string;
              settleMissing: string;
              settleRest: string;
              settleNoRule: string;
              lateMinutes: string;
              earlyMinutes: string;
              workHours: string;
              addTitle: string;
              editTitle: string;
              exportTitle: string;
              importTitle: string;
              importHint: string;
              importFileRequired: string;
              importSuccess: string;
              settleTitle: string;
              settleScope: string;
              settleScopeHint: string;
              settleRangeRequired: string;
              settleSuccess: string;
              form: {
                name: string;
                clockInStatus: string;
                location: string;
                employeeUuid: string;
                punchDate: string;
              };
            };
            overtimeRule: {
              title: string;
              ruleName: string;
              workday: string;
              restday: string;
              holiday: string;
              dayPolicy: string;
              approvalRequired: string;
              punchRequired: string;
              compMultiplier: string;
              multiplierTip: string;
              multiplierValue: string;
              needApproval: string;
              noApproval: string;
              needPunch: string;
              noPunch: string;
              minUnitMinutes: string;
              minUnitHalfHour: string;
              minUnitHour: string;
              monthlyLimitHours: string;
              monthlyLimitTip: string;
              noLimit: string;
              usedGroups: string;
              boundGroupCount: string;
              globalFallback: string;
              addRule: string;
              editRule: string;
              enable: string;
              disable: string;
              status: string;
              remark: string;
              form: {
                ruleName: string;
                status: string;
                remark: string;
              };
            };
            scheduleCycle: {
              title: string;
              cycleName: string;
              shiftSequence: string;
              cycleDays: string;
              cycleDaysTip: string;
              sequenceSummary: string;
              usedGroups: string;
              boundGroupCount: string;
              noBoundGroup: string;
              addCycle: string;
              editCycle: string;
              enable: string;
              disable: string;
              status: string;
              remark: string;
              form: {
                cycleName: string;
                shiftSequence: string;
                usedGroups: string;
                status: string;
                remark: string;
              };
            };
            scheduleBoard: {
              title: string;
              searchTitle: string;
              group: string;
              groupPlaceholder: string;
              groupLabel: string;
              month: string;
              employeeName: string;
              employeeNo: string;
              orgName: string;
              selectGroupFirst: string;
              cellHint: string;
              publishedLocked: string;
              publishedCount: string;
              generate: string;
              copyWeek: string;
              import: string;
              export: string;
              publish: string;
              publishSuccess: string;
              generateSuccess: string;
              copySuccess: string;
              importSuccess: string;
              importFileRequired: string;
              importHint: string;
              cycle: string;
              cycleDaysUnit: string;
              cycleRequired: string;
              anchorDate: string;
              anchorHintLabel: string;
              anchorHint: string;
              overwrite: string;
              sourceWeek: string;
              targetWeek: string;
              weekHintLabel: string;
              weekHint: string;
              weekRequired: string;
            };
          };
          employeeRelation: {
            billStatus: {
              title: string;
              unsubmitted: string;
              approving: string;
              approved: string;
              rejected: string;
              revoked: string;
            };
            transfer: {
              title: string;
              billNo: string;
              employeeNo: string;
              employeeName: string;
              changeType: {
                title: string;
                post: string;
                org: string;
                salary: string;
                company: string;
                other: string;
              };
              originalCompany: string;
              originalOrg: string;
              originalPosition: string;
              originalPost: string;
              newCompany: string;
              newOrg: string;
              newPosition: string;
              newPost: string;
              newSalary: string;
              transferReason: string;
              effectiveImmediately: string;
              effectiveDate: string;
              applied: string;
              creatorName: string;
              createTime: string;
              yes: string;
              no: string;
              submit: string;
              approve: string;
              reject: string;
              revoke: string;
              records: string;
              noRecords: string;
              confirmSubmit: string;
              confirmApprove: string;
              confirmReject: string;
              confirmRevoke: string;
              addTransfer: string;
              editTransfer: string;
              sectionEmployee: string;
              sectionOriginal: string;
              sectionNew: string;
              sectionEffect: string;
              form: {
                billNo: string;
                status: string;
                employeeNo: string;
                employeeName: string;
                changeType: string;
                org: string;
                transferReason: string;
                effectiveDateRequired: string;
              };
            };
            regularization: {
              title: string;
              billNo: string;
              employeeNo: string;
              employeeName: string;
              selectEmployee: string;
              orgUuid: string;
              companyUuid: string;
              position: string;
              post: string;
              personStatus: string;
              hireDate: string;
              formalDate: string;
              expectedFormalDate: string;
              remark: string;
              creatorName: string;
              createTime: string;
              submit: string;
              approve: string;
              reject: string;
              revoke: string;
              records: string;
              noRecords: string;
              confirmSubmit: string;
              confirmApprove: string;
              confirmReject: string;
              confirmRevoke: string;
              addRegularization: string;
              editRegularization: string;
              sectionEmployee: string;
              sectionSnapshot: string;
              sectionRegular: string;
              form: {
                billNo: string;
                status: string;
                employeeNo: string;
                selectEmployee: string;
                remark: string;
                none: string;
              };
            };
            resignation: {
              title: string;
              billNo: string;
              employeeNo: string;
              employeeName: string;
              selectEmployee: string;
              orgUuid: string;
              companyUuid: string;
              position: string;
              post: string;
              resignType: {
                title: string;
                voluntary: string;
                involuntary: string;
                contract_end: string;
                retirement: string;
              };
              resignReason: string;
              applyDate: string;
              resignDate: string;
              lastWorkDate: string;
              handoverPerson: string;
              salarySettleDate: string;
              creatorName: string;
              createTime: string;
              submit: string;
              approve: string;
              reject: string;
              revoke: string;
              records: string;
              noRecords: string;
              confirmSubmit: string;
              confirmApprove: string;
              confirmReject: string;
              confirmRevoke: string;
              addResignation: string;
              editResignation: string;
              sectionEmployee: string;
              sectionSnapshot: string;
              sectionResign: string;
              form: {
                billNo: string;
                status: string;
                employeeNo: string;
                selectEmployee: string;
                remark: string;
                none: string;
                resignType: string;
                handoverPerson: string;
              };
            };
            entry: {
              title: string;
              billNo: string;
              billStatus: {
                title: string;
              };
              name: string;
              phone: string;
              email: string;
              gender: string;
              birthday: string;
              nation: string;
              politicalStatus: string;
              maritalStatus: string;
              cultureLevel: string;
              nativePlace: string;
              address: string;
              householdAddress: string;
              emergencyContact: string;
              emergencyPhone: string;
              bankName: string;
              bankAccount: string;
              avatar: string;
              position: string;
              post: string;
              probationMonths: string;
              orgUuid: string;
              companyUuid: string;
              hireDate: string;
              expectFormalDate: string;
              attachmentDesc: string;
              remark: string;
              archived: string;
              archivedYes: string;
              archivedNo: string;
              creatorName: string;
              createTime: string;
              submit: string;
              approve: string;
              reject: string;
              revoke: string;
              records: string;
              noRecords: string;
              confirmSubmit: string;
              confirmApprove: string;
              confirmReject: string;
              confirmRevoke: string;
              addEntry: string;
              editEntry: string;
              tabBasic: string;
              tabWorks: string;
              tabEducations: string;
              tabFamilies: string;
              tabAttachments: string;
              sectionBasic: string;
              addRow: string;
              noRows: string;
              workCompany: string;
              workPosition: string;
              school: string;
              major: string;
              education: string;
              familyName: string;
              relation: string;
              familyPhone: string;
              workUnit: string;
              startDate: string;
              endDate: string;
              form: {
                name: string;
                gender: string;
                birthday: string;
                phone: string;
                email: string;
                nation: string;
                politicalStatus: string;
                maritalStatus: string;
                cultureLevel: string;
                nativePlace: string;
                address: string;
                householdAddress: string;
                emergencyContact: string;
                emergencyPhone: string;
                bankName: string;
                bankAccount: string;
                avatar: string;
                position: string;
                post: string;
                orgUuid: string;
                companyUuid: string;
                expectFormalDate: string;
                attachmentDesc: string;
                remark: string;
                billNo: string;
                status: string;
              };
            };
          };
          reward: {
            title: string;
            billNo: string;
            billStatus: {
              title: string;
            };
            employeeNo: string;
            employeeName: string;
            rewardType: {
              title: string;
              reward: string;
              punishment: string;
            };
            rewardLevel: {
              title: string;
              company: string;
              dept: string;
              notice: string;
              other: string;
            };
            docNo: string;
            occurDate: string;
            amount: string;
            reason: string;
            salaryMonth: string;
            synced: string;
            creatorName: string;
            createTime: string;
            yes: string;
            no: string;
            submit: string;
            approve: string;
            reject: string;
            revoke: string;
            records: string;
            noRecords: string;
            confirmSubmit: string;
            confirmApprove: string;
            confirmReject: string;
            confirmRevoke: string;
            addReward: string;
            editReward: string;
            sectionEmployee: string;
            sectionReward: string;
            form: {
              billNo: string;
              status: string;
              rewardType: string;
              rewardLevel: string;
              employeeNo: string;
              employeeName: string;
              docNo: string;
              salaryMonth: string;
              reason: string;
            };
          };
          requirement: {
            title: string;
            billNo: string;
            billStatus: {
              title: string;
            };
            postName: string;
            postCategory: {
              title: string;
              tech: string;
              product: string;
              design: string;
              operation: string;
              market: string;
              function: string;
              other: string;
            };
            headcount: string;
            urgency: {
              title: string;
              normal: string;
              urgent: string;
            };
            expectDate: string;
            orgName: string;
            requirementDesc: string;
            creatorName: string;
            createTime: string;
            submit: string;
            approve: string;
            reject: string;
            revoke: string;
            records: string;
            noRecords: string;
            confirmSubmit: string;
            confirmApprove: string;
            confirmReject: string;
            confirmRevoke: string;
            addRequirement: string;
            editRequirement: string;
            sectionPost: string;
            form: {
              billNo: string;
              status: string;
              postName: string;
              postCategory: string;
              urgency: string;
              org: string;
              requirementDesc: string;
            };
          };
          candidate: {
            title: string;
            name: string;
            phone: string;
            email: string;
            sourceChannel: {
              title: string;
              recommend: string;
              website: string;
              campus: string;
              headhunter: string;
              other: string;
            };
            education: {
              title: string;
              doctor: string;
              master: string;
              bachelor: string;
              associate: string;
              senior: string;
              other: string;
            };
            currentCompany: string;
            currentPosition: string;
            expectSalary: string;
            remark: string;
            status: {
              title: string;
              new: string;
              delivered: string;
              interviewing: string;
              offer: string;
              hired: string;
              rejected: string;
            };
            createTime: string;
            advanceTo: string;
            confirmAdvance: string;
            confirmReject: string;
            reject: string;
            import: string;
            importModalTitle: string;
            importSuccess: string;
            importTemplateTip: string;
            chooseFileTip: string;
            addCandidate: string;
            editCandidate: string;
            form: {
              name: string;
              phone: string;
              email: string;
              sourceChannel: string;
              education: string;
              currentCompany: string;
              currentPosition: string;
              expectSalary: string;
              remark: string;
            };
          };
          salary: {
            subject: {
              title: string;
              subjectName: string;
              orgName: string;
              payAccount: string;
              payBank: string;
              status: string;
              remark: string;
              addSubject: string;
              editSubject: string;
              form: {
                subjectName: string;
                orgName: string;
                payAccount: string;
                payBank: string;
                status: string;
                remark: string;
              };
            };
            item: {
              title: string;
              itemName: string;
              itemCode: string;
              itemType: {
                title: string;
                fixed: string;
                data: string;
                formula: string;
              };
              itemClass: {
                title: string;
                earning: string;
                deduction: string;
                cost: string;
              };
              valueMode: {
                title: string;
                manual: string;
                system: string;
                formula: string;
              };
              calcOrder: string;
              orderNum: string;
              formula: string;
              isPreset: string;
              presetYes: string;
              presetNo: string;
              status: string;
              remark: string;
              addItem: string;
              editItem: string;
              form: {
                itemName: string;
                itemCode: string;
                itemType: string;
                itemClass: string;
                valueMode: string;
                formula: string;
                status: string;
                remark: string;
              };
            };
            tax: {
              profile: {
                title: string;
                employeeName: string;
                idCard: string;
                residentType: string;
                residentTypeEnum: {
                  resident: string;
                  non_resident: string;
                };
                taxCity: string;
                defaultSpecialExtra: string;
                status: string;
                remark: string;
                addProfile: string;
                editProfile: string;
                form: {
                  employeeName: string;
                  residentType: string;
                  taxCity: string;
                  defaultSpecialExtra: string;
                  status: string;
                  remark: string;
                };
              };
              deduction: {
                title: string;
                employeeName: string;
                idCard: string;
                deductYear: string;
                deductMonth: string;
                deductTypeTitle: string;
                amount: string;
                sourceTitle: string;
                remark: string;
                maintain: string;
                maintainYear: string;
                import: string;
                importSuccess: string;
                importModalTitle: string;
                chooseFileTip: string;
                maintainModalTitle: string;
                maintainYearTitle: string;
                saved: string;
                form: {
                  employeeName: string;
                  deductYear: string;
                  amount: string;
                  months: string;
                };
                deductType: {
                  CHILDREN_EDU: string;
                  CONTINUING_EDU: string;
                  CRITICAL_ILLNESS: string;
                  HOUSE_LOAN: string;
                  HOUSE_RENT: string;
                  SUPPORT_ELDERLY: string;
                  INFANT_CARE: string;
                };
                source: {
                  manual: string;
                  import: string;
                };
              };
              opening: {
                title: string;
                employeeName: string;
                idCard: string;
                subjectName: string;
                taxYear: string;
                untilMonth: string;
                incomeTotal: string;
                specialDeductionTotal: string;
                specialExtraTotal: string;
                exemptTotal: string;
                otherDeductionTotal: string;
                taxableIncomeTotal: string;
                taxPaidTotal: string;
                remark: string;
                addOpening: string;
                editOpening: string;
                import: string;
                importSuccess: string;
                importModalTitle: string;
                chooseFileTip: string;
                form: {
                  employeeName: string;
                  subjectName: string;
                  taxYear: string;
                  untilMonth: string;
                };
              };
            };
          };
          training: {
            mySessionTitle: string;
            myCertTitle: string;
            sessionNo: string;
            sessionName: string;
            location: string;
            lecturer: string;
            hours: string;
            startTime: string;
            endTime: string;
            sessionStatusLabel: string;
            sessionStatus: {
              notStarted: string;
              ongoing: string;
              finished: string;
            };
            certName: string;
            certNo: string;
            issueTime: string;
            certStatusLabel: string;
            certStatus: {
              pending: string;
              issued: string;
              revoked: string;
            };
            scanFile: string;
            preview: string;
            noFile: string;
            planType: {
              special: string;
              newHire: string;
              safety: string;
              skill: string;
            };
            published: {
              draft: string;
              published: string;
            };
            plan: {
              title: string;
              billNo: string;
              billStatusLabel: string;
              planName: string;
              theme: string;
              planYear: string;
              planType: string;
              startDate: string;
              endDate: string;
              published: string;
              creatorName: string;
              addPlan: string;
              editPlan: string;
              submit: string;
              approve: string;
              reject: string;
              revoke: string;
              records: string;
              confirmSubmit: string;
              confirmApprove: string;
              confirmReject: string;
              confirmRevoke: string;
              form: {
                planName: string;
                theme: string;
                planYear: string;
                planType: string;
                status: string;
                published: string;
                endDateBeforeStart: string;
              };
            };
            session: {
              title: string;
              sessionName: string;
              planName: string;
              lecturer: string;
              remark: string;
              addSession: string;
              editSession: string;
              participants: string;
              finish: string;
              confirmFinish: string;
              participantsTip: string;
              form: {
                sessionName: string;
                planUuid: string;
                lecturer: string;
                location: string;
                status: string;
                participantsPlaceholder: string;
                endTimeBeforeStart: string;
              };
            };
            certTemplate: {
              title: string;
              templateName: string;
              certType: string;
              sourceLabel: string;
              noPrefix: string;
              statusLabel: string;
              remark: string;
              addTemplate: string;
              editTemplate: string;
              enable: string;
              disable: string;
              confirmEnable: string;
              confirmDisable: string;
              source: {
                system: string;
                org: string;
              };
              status: {
                enabled: string;
                disabled: string;
              };
              form: {
                templateName: string;
                certType: string;
                source: string;
                noPrefix: string;
              };
            };
            certificate: {
              title: string;
              student: string;
              template: string;
              sourceLabel: string;
              source: {
                session: string;
                import: string;
              };
              session: string;
              batchIssue: string;
              batchIssueDone: string;
              batchIssueSkipped: string;
              batchIssueTip: string;
              importExternal: string;
              issue: string;
              revoke: string;
              upload: string;
              chooseFile: string;
              confirmIssue: string;
              confirmRevoke: string;
              form: {
                session: string;
                template: string;
                student: string;
                certName: string;
                certType: string;
                certNo: string;
                status: string;
                source: string;
              };
            };
          };
          employee: {
            title: string;
            employeeNo: string;
            name: string;
            genderLabel: string;
            gender: {
              male: string;
              female: string;
            };
            orgName: string;
            companyName: string;
            position: string;
            post: string;
            personStatusLabel: string;
            personStatus: {
              regular: string;
              probation: string;
              intern: string;
              temp: string;
              resigned: string;
              retired: string;
            };
            hireDate: string;
            formalDate: string;
            phone: string;
            email: string;
            userStatusLabel: string;
            userStatus: {
              created: string;
              none: string;
            };
            birthday: string;
            nation: string;
            politicalStatus: string;
            maritalStatusLabel: string;
            maritalStatus: {
              unmarried: string;
              married: string;
              divorced: string;
              widowed: string;
            };
            nativePlace: string;
            bloodType: string;
            idCard: string;
            householdAddress: string;
            address: string;
            emergencyContact: string;
            emergencyPhone: string;
            bankName: string;
            bankAccount: string;
            avatar: string;
            remark: string;
            createTime: string;
            addEmployee: string;
            editEmployee: string;
            importEmployee: string;
            exportEmployee: string;
            generateUser: string;
            generateUserSelectTip: string;
            generateUserSuccess: string;
            generateUserAllSkipped: string;
            importSuccess: string;
            chooseFileTip: string;
            importModalTitle: string;
            tabBasic: string;
            tabEducation: string;
            tabWork: string;
            tabFamily: string;
            addEducationRow: string;
            noEducationRows: string;
            addWorkRow: string;
            noWorkRows: string;
            addFamilyRow: string;
            noFamilyRows: string;
            eduSchool: string;
            eduMajor: string;
            eduEducation: string;
            eduStartDate: string;
            eduEndDate: string;
            workCompany: string;
            workPosition: string;
            workStartDate: string;
            workEndDate: string;
            familyName: string;
            familyRelation: string;
            familyPhone: string;
            familyWorkUnit: string;
            form: {
              employeeNo: string;
              employeeNoAuto: string;
              name: string;
              org: string;
              company: string;
              position: string;
              post: string;
              personStatus: string;
              phone: string;
              email: string;
              birthday: string;
              nation: string;
              politicalStatus: string;
              maritalStatus: string;
              nativePlace: string;
              bloodType: string;
              idCard: string;
              householdAddress: string;
              address: string;
              emergencyContact: string;
              emergencyPhone: string;
              bankName: string;
              bankAccount: string;
              remark: string;
            };
            contractExpiring: {
              title: string;
              employeeNo: string;
              employeeName: string;
              selectEmployee: string;
              orgName: string;
              contractNo: string;
              startDate: string;
              endDate: string;
              daysToExpiry: string;
              renewCount: string;
              status: string;
              stateNormal: string;
              stateWarning: string;
              stateExpired: string;
              stateTerminated: string;
              statusActive: string;
              statusExpired: string;
              statusTerminated: string;
              renew: string;
              renewSuccess: string;
              confirmRenew: string;
              addContract: string;
              editContract: string;
              form: {
                employeeName: string;
                employeeNo: string;
                status: string;
                selectEmployee: string;
                contractNo: string;
                remark: string;
              };
            };
            contractRenewable: {
              title: string;
              employeeNo: string;
              employeeName: string;
              orgName: string;
              startDate: string;
              endDate: string;
              daysToExpiry: string;
              renewCount: string;
              status: string;
              stateNormal: string;
              stateWarning: string;
              stateExpired: string;
              stateTerminated: string;
              renew: string;
              renewSuccess: string;
              confirmRenew: string;
              form: {
                employeeName: string;
                employeeNo: string;
                status: string;
              };
            };
            certExpiring: {
              title: string;
              employeeNo: string;
              employeeName: string;
              selectEmployee: string;
              orgName: string;
              certType: string;
              certNo: string;
              issueDate: string;
              expiryDate: string;
              daysToExpiry: string;
              status: string;
              stateNormal: string;
              stateWarning: string;
              stateExpired: string;
              addCert: string;
              editCert: string;
              form: {
                selectEmployee: string;
                certType: string;
                certNo: string;
              };
            };
          };
        };
      };
      form: {
        required: string;
        username: FormMsg;
        phone: FormMsg;
        pwd: FormMsg;
        confirmPwd: FormMsg;
        code: FormMsg;
        email: FormMsg;
      };
      dropdown: Record<Global.DropdownKey, string>;
      icon: {
        themeConfig: string;
        themeSchema: string;
        lang: string;
        fullscreen: string;
        fullscreenExit: string;
        reload: string;
        collapse: string;
        expand: string;
        pin: string;
        unpin: string;
      };
      datatable: {
        itemCount: string;
      };
    };

    type GetI18nKey<T extends Record<string, unknown>, K extends keyof T = keyof T> = K extends string
      ? T[K] extends Record<string, unknown>
        ? `${K}.${GetI18nKey<T[K]>}`
        : K
      : never;

    type I18nKey = GetI18nKey<Schema>;

    type TranslateOptions<Locales extends string> = import('vue-i18n').TranslateOptions<Locales>;

    interface $T {
      (key: I18nKey): string;
      (key: I18nKey, plural: number, options?: TranslateOptions<LangType>): string;
      (key: I18nKey, defaultMsg: string, options?: TranslateOptions<I18nKey>): string;
      (key: I18nKey, list: unknown[], options?: TranslateOptions<I18nKey>): string;
      (key: I18nKey, list: unknown[], plural: number): string;
      (key: I18nKey, list: unknown[], defaultMsg: string): string;
      (key: I18nKey, named: Record<string, unknown>, options?: TranslateOptions<LangType>): string;
      (key: I18nKey, named: Record<string, unknown>, plural: number): string;
      (key: I18nKey, named: Record<string, unknown>, defaultMsg: string): string;
    }
  }

  /** Service namespace */
  namespace Service {
    /** Other baseURL key */
    type OtherBaseURLKey = 'demo';

    interface ServiceConfigItem {
      /** The backend service base url */
      baseURL: string;
      /** The proxy pattern of the backend service base url */
      proxyPattern: string;
    }

    interface OtherServiceConfigItem extends ServiceConfigItem {
      key: OtherBaseURLKey;
    }

    /** The backend service config */
    interface ServiceConfig extends ServiceConfigItem {
      /** Other backend service config */
      other: OtherServiceConfigItem[];
    }

    interface SimpleServiceConfig extends Pick<ServiceConfigItem, 'baseURL'> {
      other: Record<OtherBaseURLKey, string>;
    }

    /** The backend service response data */
    type Response<T = unknown> = {
      /** The backend service response code */
      code: string;
      /** The backend service response message */
      msg: string;
      /** The backend service response data */
      data: T;
    };

    /** The demo backend service response data */
    type DemoResponse<T = unknown> = {
      /** The backend service response code */
      status: string;
      /** The backend service response message */
      message: string;
      /** The backend service response data */
      result: T;
    };
  }
}
