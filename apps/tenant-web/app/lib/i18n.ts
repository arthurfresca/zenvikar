export type Locale = "en" | "pt";

export type Translations = {
  // Sidebar / layout
  nav_navigate: string;
  nav_overview: string;
  nav_settings: string;
  nav_services: string;
  nav_team: string;
  nav_bookings: string;
  nav_customers: string;
  nav_studio: string;
  nav_switch_workspace: string;
  contact: string;
  live_badge: string;
  hidden_badge: string;
  sign_out: string;
  // Overview
  overview_eyebrow: string;
  service_studio: string;
  apply: string;
  booking_page_live: string;
  upcoming: string;
  no_upcoming: string;
  all_bookings_link: string;
  needs_attention: string;
  view_all_link: string;
  // Metric cards
  metric_live_services: string;
  metric_team_members: string;
  metric_confirmed: string;
  metric_pending: string;
  total_bookings: string;
  cancelled_count: string;
  assigned_count: string;
  live_lower: string;
  hidden_lower: string;
  active_lower: string;
  // Notifications
  action_completed: string;
  could_not_complete: string;
  // Settings sections
  settings_brand: string;
  settings_brand_desc: string;
  settings_contact: string;
  settings_contact_desc: string;
  settings_preferences: string;
  settings_preferences_desc: string;
  settings_booking_page_desc: string;
  settings_booking_page_hint: string;
  // Settings fields
  identity_eyebrow: string;
  brand_settings_title: string;
  display_name: string;
  logo_url: string;
  primary_color: string;
  primary_color_hint: string;
  secondary_color: string;
  secondary_color_hint: string;
  accent_color: string;
  accent_color_hint: string;
  contact_phone: string;
  contact_email: string;
  address: string;
  currency: string;
  timezone: string;
  locale: string;
  booking_page: string;
  enabled_for_customers: string;
  save_settings: string;
  // Services
  catalog_eyebrow: string;
  add_service_title: string;
  service_name: string;
  description: string;
  description_placeholder: string;
  duration_min: string;
  buffer_before: string;
  buffer_after: string;
  visibility: string;
  publish_immediately: string;
  create_service: string;
  service_roster: string;
  no_services: string;
  no_services_title: string;
  no_services_body: string;
  cancel: string;
  assigned_specialists: string;
  no_specialists_yet: string;
  add_specialist: string;
  team_member: string;
  choose_team_member: string;
  all_members_assigned: string;
  price_label: string;
  internal_note: string;
  internal_note_placeholder: string;
  edit_service_details: string;
  visible_on_booking_page: string;
  save_changes: string;
  duplicate: string;
  delete_service: string;
  calendar_controls: string;
  hide_service: string;
  publish_service: string;
  remove: string;
  // Workspace info
  workspace: string;
  timezone_label: string;
  currency_label: string;
  locale_label: string;
  // Team
  team_eyebrow: string;
  staff_roles: string;
  member_singular: string;
  members_plural: string;
  service_singular_lc: string;
  services_plural_lc: string;
  // Status badges
  status_confirmed: string;
  status_cancelled: string;
  status_pending: string;
  // Customers
  customers_eyebrow: string;
  frequent_guests: string;
  no_customers_yet: string;
  visits: string;
  lifetime: string;
  top_guests_subtitle: string;
  select_customer_hint: string;
  booking_history: string;
  // Bookings panel
  bookings_eyebrow: string;
  bookings_ops_title: string;
  search_placeholder: string;
  no_bookings_filter: string;
  booking_singular: string;
  bookings_plural_lc: string;
  select_booking_hint: string;
  service_label: string;
  specialist_label: string;
  start_label: string;
  end_label: string;
  value_label: string;
  confirm_booking: string;
  mark_pending: string;
  cancel_booking: string;
  filter_all: string;
  filter_pending: string;
  filter_confirmed: string;
  filter_cancelled: string;
  // Schedule editor
  weekly_hours: string;
  template_weekdays: string;
  template_extended: string;
  template_weekend: string;
  apply_template: string;
  copy_from_specialist: string;
  copy: string;
  closed: string;
  open_label: string;
  close_label: string;
  save_prefix: string;
  blocked_dates: string;
  date_from: string;
  date_to: string;
  reason_placeholder: string;
  block_date_range: string;
  no_blocked_dates: string;
  calendar_badge: string;
  days_short: [string, string, string, string, string, string, string];
  // Intl locale tag for number/date formatting
  intl_locale: string;
};

const en: Translations = {
  nav_navigate: "Navigate",
  nav_overview: "Overview",
  nav_settings: "Settings",
  nav_services: "Services",
  nav_team: "Team",
  nav_bookings: "Bookings",
  nav_customers: "Customers",
  nav_studio: "Studio",
  nav_switch_workspace: "Switch workspace",
  contact: "Contact",
  live_badge: "Live",
  hidden_badge: "Hidden",
  sign_out: "Sign out",
  overview_eyebrow: "Overview",
  service_studio: "Service Studio",
  apply: "Apply",
  booking_page_live: "Booking page live",
  upcoming: "Upcoming",
  no_upcoming: "No upcoming bookings in the selected range.",
  all_bookings_link: "All bookings →",
  needs_attention: "Needs attention",
  view_all_link: "View all →",
  metric_live_services: "Live services",
  metric_team_members: "Team members",
  metric_confirmed: "Confirmed",
  metric_pending: "Pending",
  total_bookings: "total bookings",
  cancelled_count: "cancelled",
  assigned_count: "assigned",
  live_lower: "live",
  hidden_lower: "hidden",
  active_lower: "active",
  action_completed: "Completed",
  could_not_complete: "Could not complete",
  settings_brand: "Brand & identity",
  settings_brand_desc: "How your workspace appears to customers",
  settings_contact: "Contact information",
  settings_contact_desc: "How customers can reach you",
  settings_preferences: "Preferences",
  settings_preferences_desc: "Locale and regional settings",
  settings_booking_page_desc: "Control whether customers can book your services online",
  settings_booking_page_hint: "Allow customers to view and book your services",
  identity_eyebrow: "Identity",
  brand_settings_title: "Brand & booking settings",
  display_name: "Display name",
  logo_url: "Logo URL",
  primary_color: "Primary color",
  primary_color_hint: "Used for buttons and main actions on your booking page",
  secondary_color: "Secondary color",
  secondary_color_hint: "Used for backgrounds and secondary UI elements",
  accent_color: "Accent color",
  accent_color_hint: "Used for highlights, links and interactive elements",
  contact_phone: "Contact phone",
  contact_email: "Contact email",
  address: "Address",
  currency: "Currency",
  timezone: "Timezone",
  locale: "Locale",
  booking_page: "Booking page",
  enabled_for_customers: "Enabled for customers",
  save_settings: "Save settings",
  catalog_eyebrow: "Catalog",
  add_service_title: "Add a new service",
  service_name: "Service name",
  description: "Description",
  description_placeholder: "What customers get in this appointment",
  duration_min: "Duration (min)",
  buffer_before: "Buffer before (min)",
  buffer_after: "Buffer after (min)",
  visibility: "Visibility",
  publish_immediately: "Publish immediately",
  create_service: "Create service",
  service_roster: "Service roster",
  no_services: "No services yet — add your first service above.",
  no_services_title: "No services yet",
  no_services_body: "Create your first service to start accepting bookings.",
  cancel: "Cancel",
  assigned_specialists: "Assigned specialists",
  no_specialists_yet: "No specialists assigned yet.",
  add_specialist: "Add specialist",
  team_member: "Team member",
  choose_team_member: "Choose a team member",
  all_members_assigned: "All members assigned",
  price_label: "Price",
  internal_note: "Internal note",
  internal_note_placeholder: "Senior stylist, premium pricing…",
  edit_service_details: "Edit service details",
  visible_on_booking_page: "Visible on booking page",
  save_changes: "Save changes",
  duplicate: "Duplicate",
  delete_service: "Delete service",
  calendar_controls: "Calendar controls",
  hide_service: "Hide service",
  publish_service: "Publish service",
  remove: "Remove",
  workspace: "Workspace",
  timezone_label: "Timezone",
  currency_label: "Currency",
  locale_label: "Locale",
  team_eyebrow: "Team",
  staff_roles: "Staff & roles",
  member_singular: "member",
  members_plural: "members",
  service_singular_lc: "service",
  services_plural_lc: "services",
  status_confirmed: "Confirmed",
  status_cancelled: "Cancelled",
  status_pending: "Pending",
  customers_eyebrow: "Customers",
  frequent_guests: "Frequent guests",
  no_customers_yet: "Data will appear once bookings arrive.",
  visits: "visits",
  lifetime: "lifetime",
  top_guests_subtitle: "Top guests by visit count",
  select_customer_hint: "Select a customer to view details",
  booking_history: "History",
  bookings_eyebrow: "Bookings",
  bookings_ops_title: "Operations",
  search_placeholder: "Search customer, service…",
  no_bookings_filter: "No bookings match this filter.",
  booking_singular: "booking",
  bookings_plural_lc: "bookings",
  select_booking_hint: "Select a booking to view details",
  service_label: "Service",
  specialist_label: "Specialist",
  start_label: "Start",
  end_label: "End",
  value_label: "Value",
  confirm_booking: "Confirm",
  mark_pending: "Mark pending",
  cancel_booking: "Cancel",
  filter_all: "all",
  filter_pending: "pending",
  filter_confirmed: "confirmed",
  filter_cancelled: "cancelled",
  weekly_hours: "Weekly hours",
  template_weekdays: "Weekdays 09:00–18:00",
  template_extended: "Extended 08:00–20:00",
  template_weekend: "Weekend 10:00–16:00",
  apply_template: "Apply",
  copy_from_specialist: "Copy from specialist…",
  copy: "Copy",
  closed: "Closed",
  open_label: "Open",
  close_label: "Close",
  save_prefix: "Save",
  blocked_dates: "Blocked dates",
  date_from: "From",
  date_to: "To",
  reason_placeholder: "Reason — vacation, training, medical…",
  block_date_range: "Block date range",
  no_blocked_dates: "No blocked dates",
  calendar_badge: "Calendar",
  days_short: ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"],
  intl_locale: "en-US",
};

const pt: Translations = {
  nav_navigate: "Navegar",
  nav_overview: "Visão geral",
  nav_settings: "Configurações",
  nav_services: "Serviços",
  nav_team: "Equipe",
  nav_bookings: "Agendamentos",
  nav_customers: "Clientes",
  nav_studio: "Studio",
  nav_switch_workspace: "Trocar workspace",
  contact: "Contato",
  live_badge: "Ativo",
  hidden_badge: "Oculto",
  sign_out: "Sair",
  overview_eyebrow: "Visão geral",
  service_studio: "Estúdio de Serviços",
  apply: "Aplicar",
  booking_page_live: "Página de agendamento ativa",
  upcoming: "Próximos",
  no_upcoming: "Nenhum agendamento confirmado no período selecionado.",
  all_bookings_link: "Todos os agendamentos →",
  needs_attention: "Requer atenção",
  view_all_link: "Ver todos →",
  metric_live_services: "Serviços ativos",
  metric_team_members: "Membros da equipe",
  metric_confirmed: "Confirmados",
  metric_pending: "Pendentes",
  total_bookings: "agendamentos no total",
  cancelled_count: "cancelados",
  assigned_count: "atribuídos",
  live_lower: "ativo",
  hidden_lower: "oculto",
  active_lower: "ativo",
  action_completed: "Concluído",
  could_not_complete: "Não foi possível concluir",
  settings_brand: "Marca e identidade",
  settings_brand_desc: "Como seu workspace aparece para os clientes",
  settings_contact: "Informações de contato",
  settings_contact_desc: "Como os clientes podem entrar em contato",
  settings_preferences: "Preferências",
  settings_preferences_desc: "Configurações de idioma e região",
  settings_booking_page_desc: "Controle se os clientes podem agendar seus serviços online",
  settings_booking_page_hint: "Permitir que os clientes visualizem e agendem seus serviços",
  identity_eyebrow: "Identidade",
  brand_settings_title: "Marca e configurações de agendamento",
  display_name: "Nome de exibição",
  logo_url: "URL do logotipo",
  primary_color: "Cor primária",
  primary_color_hint: "Usada para botões e ações principais na sua página de agendamento",
  secondary_color: "Cor secundária",
  secondary_color_hint: "Usada para planos de fundo e elementos secundários da interface",
  accent_color: "Cor de destaque",
  accent_color_hint: "Usada para destaques, links e elementos interativos",
  contact_phone: "Telefone de contato",
  contact_email: "E-mail de contato",
  address: "Endereço",
  currency: "Moeda",
  timezone: "Fuso horário",
  locale: "Idioma",
  booking_page: "Página de agendamento",
  enabled_for_customers: "Habilitado para clientes",
  save_settings: "Salvar configurações",
  catalog_eyebrow: "Catálogo",
  add_service_title: "Adicionar um novo serviço",
  service_name: "Nome do serviço",
  description: "Descrição",
  description_placeholder: "O que os clientes recebem neste agendamento",
  duration_min: "Duração (min)",
  buffer_before: "Intervalo antes (min)",
  buffer_after: "Intervalo depois (min)",
  visibility: "Visibilidade",
  publish_immediately: "Publicar imediatamente",
  create_service: "Criar serviço",
  service_roster: "Lista de serviços",
  no_services: "Nenhum serviço ainda — adicione seu primeiro serviço acima.",
  no_services_title: "Nenhum serviço ainda",
  no_services_body: "Crie seu primeiro serviço para começar a receber agendamentos.",
  cancel: "Cancelar",
  assigned_specialists: "Especialistas atribuídos",
  no_specialists_yet: "Nenhum especialista atribuído ainda.",
  add_specialist: "Adicionar especialista",
  team_member: "Membro da equipe",
  choose_team_member: "Escolha um membro da equipe",
  all_members_assigned: "Todos os membros foram atribuídos",
  price_label: "Preço",
  internal_note: "Nota interna",
  internal_note_placeholder: "Estilista sênior, preços premium…",
  edit_service_details: "Editar detalhes do serviço",
  visible_on_booking_page: "Visível na página de agendamento",
  save_changes: "Salvar alterações",
  duplicate: "Duplicar",
  delete_service: "Excluir serviço",
  calendar_controls: "Controles de calendário",
  hide_service: "Ocultar serviço",
  publish_service: "Publicar serviço",
  remove: "Remover",
  workspace: "Workspace",
  timezone_label: "Fuso horário",
  currency_label: "Moeda",
  locale_label: "Idioma",
  team_eyebrow: "Equipe",
  staff_roles: "Funcionários e funções",
  member_singular: "membro",
  members_plural: "membros",
  service_singular_lc: "serviço",
  services_plural_lc: "serviços",
  status_confirmed: "Confirmado",
  status_cancelled: "Cancelado",
  status_pending: "Pendente",
  customers_eyebrow: "Clientes",
  frequent_guests: "Clientes frequentes",
  no_customers_yet: "Os dados aparecerão quando os agendamentos chegarem.",
  visits: "visitas",
  lifetime: "total",
  top_guests_subtitle: "Principais clientes por número de visitas",
  select_customer_hint: "Selecione um cliente para ver os detalhes",
  booking_history: "Histórico",
  bookings_eyebrow: "Agendamentos",
  bookings_ops_title: "Operações",
  search_placeholder: "Buscar cliente, serviço…",
  no_bookings_filter: "Nenhum agendamento corresponde a este filtro.",
  booking_singular: "agendamento",
  bookings_plural_lc: "agendamentos",
  select_booking_hint: "Selecione um agendamento para ver os detalhes",
  service_label: "Serviço",
  specialist_label: "Especialista",
  start_label: "Início",
  end_label: "Fim",
  value_label: "Valor",
  confirm_booking: "Confirmar",
  mark_pending: "Marcar como pendente",
  cancel_booking: "Cancelar",
  filter_all: "todos",
  filter_pending: "pendente",
  filter_confirmed: "confirmado",
  filter_cancelled: "cancelado",
  weekly_hours: "Horário semanal",
  template_weekdays: "Dias úteis 09:00–18:00",
  template_extended: "Estendido 08:00–20:00",
  template_weekend: "Fim de semana 10:00–16:00",
  apply_template: "Aplicar",
  copy_from_specialist: "Copiar de especialista…",
  copy: "Copiar",
  closed: "Fechado",
  open_label: "Abertura",
  close_label: "Fechamento",
  save_prefix: "Salvar",
  blocked_dates: "Datas bloqueadas",
  date_from: "De",
  date_to: "Até",
  reason_placeholder: "Motivo — férias, treinamento, médico…",
  block_date_range: "Bloquear período",
  no_blocked_dates: "Nenhuma data bloqueada",
  calendar_badge: "Calendário",
  days_short: ["Dom", "Seg", "Ter", "Qua", "Qui", "Sex", "Sáb"],
  intl_locale: "pt-BR",
};

export function getTranslations(locale: string): Translations {
  return locale === "pt" ? pt : en;
}
