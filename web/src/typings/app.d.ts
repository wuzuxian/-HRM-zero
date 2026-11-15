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
          transfer: {
            add: string;
            edit: string;
            delete: string;
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
          employee: {
            add: string;
            edit: string;
            delete: string;
            import: string;
            export: string;
            generateuser: string;
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
